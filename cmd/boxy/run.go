package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arnab2001/boxy/internal/client"
	"github.com/arnab2001/boxy/internal/cni"
	console "github.com/containerd/console"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/oci"
	refdocker "github.com/containerd/containerd/reference/docker"
	"github.com/spf13/cobra"
)

var portFlags []string

func init() {
	cmd := &cobra.Command{
		Use:   "run --name <ctr> <image> [cmd...]",
		Short: "Run a container (interactive by default, -d for detached)",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runE,
	}
	cmd.Flags().String("name", "", "container name (required)")
	cmd.Flags().BoolP("detach", "d", false, "run in background (no TTY)")
	cmd.Flags().StringSliceVarP(&portFlags, "publish", "p", nil, "HOST:CONT[,PROTO]")
	cmd.MarkFlagRequired("name")
	rootCmd.AddCommand(cmd)
}

func runE(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	detach, _ := cmd.Flags().GetBool("detach")

	// Parse port flags
	portMappings, err := ParsePorts(portFlags)
	if err != nil {
		return err
	}
	if len(portMappings) > 0 {
		fmt.Printf("Parsed port mappings: %+v\n", portMappings)
		
		// Check for port conflicts
		if err := checkPortConflicts(portMappings); err != nil {
			return err
		}
	}

	// Initialize CNI client if port mappings are specified
	var cniClient *cni.Client
	if len(portMappings) > 0 {
		cniClient, err = cni.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize CNI: %v", err)
		}
	}

	// ── normalise reference ────────────────────────────────────
	named, err := refdocker.ParseDockerRef(args[0])
	if err != nil {
		return err
	}
	ref := named.String()

	ctx := client.Default()
	c, err := client.Instance()
	if err != nil {
		return err
	}

	// ── ensure image exists (auto-pull) ────────────────────────
	img, err := c.GetImage(ctx, ref)
	if err != nil {
		if errdefs.IsNotFound(err) {
			fmt.Printf("⟳ pulling %s …\n", ref)
			img, err = c.Pull(ctx, ref, containerd.WithPullUnpack)
			if err != nil {
				return err
			}
			fmt.Printf("✔ pulled %s\n", ref)
		} else {
			return err
		}
	}

	// ── build OCI spec ─────────────────────────────────────────
	specOpts := []oci.SpecOpts{oci.WithImageConfig(img)}
	if !detach {
		specOpts = append(specOpts, oci.WithTTY)
	}
	if len(args) > 1 {
		specOpts = append(specOpts, oci.WithProcessArgs(args[1:]...))
	}

	cont, err := c.NewContainer(ctx, name,
		containerd.WithNewSnapshot(name+"-snap", img),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return err
	}

	// choose IO mode
	var creator cio.Creator
	if detach {
		creator = cio.NullIO
	} else {
		creator = cio.NewCreator(cio.WithStreams(os.Stdin, os.Stdout, os.Stderr), cio.WithTerminal)
	}

	task, err := cont.NewTask(ctx, creator)
	if err != nil {
		return err
	}
	if err := task.Start(ctx); err != nil {
		return err
	}

	fmt.Printf("▶︎ started %s (PID %d)\n", name, task.Pid())

	// ── CNI network setup ──────────────────────────────────────
	if cniClient != nil && len(portMappings) > 0 {
		netnsPath := fmt.Sprintf("/proc/%d/ns/net", task.Pid())
		
		// Convert port mappings to CNI format
		var cniPortMappings []cni.PortMapping
		for _, pm := range portMappings {
			cniPortMappings = append(cniPortMappings, cni.PortMapping{
				HostPort:      pm.HostPort,
				ContainerPort: pm.ContainerPort,
				Protocol:      pm.Protocol,
				HostIP:        pm.HostIP,
			})
		}

		result, err := cniClient.SetupNetwork(ctx, name, netnsPath, cniPortMappings)
		if err != nil {
			// Clean up task if network setup fails
			task.Kill(ctx, syscall.SIGKILL)
			task.Delete(ctx)
			cont.Delete(ctx, containerd.WithSnapshotCleanup)
			return fmt.Errorf("failed to setup network: %v", err)
		}

		fmt.Printf("✔ network configured with IP: %s\n", result.Interfaces["eth0"].IPConfigs[0].IP.String())
	}

	if detach {
		// background mode returns immediately
		return nil
	}

	// ── interactive TTY: raw mode + resize forwarding ──────────
	if cons := console.Current(); cons != nil {
		if err := cons.SetRaw(); err != nil {
			return err
		}
		defer cons.Reset()

		if sz, err := cons.Size(); err == nil { // initial resize
			_ = task.Resize(ctx, uint32(sz.Width), uint32(sz.Height))
		}
		go func() { // later resizes
			winsz := make(chan os.Signal, 1)
			signal.Notify(winsz, syscall.SIGWINCH)
			for range winsz {
				if s, err := cons.Size(); err == nil {
					_ = task.Resize(ctx, uint32(s.Width), uint32(s.Height))
				}
			}
		}()
	}

	// forward Ctrl-C / TERM into the container
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigCh
		_ = task.Kill(ctx, s.(syscall.Signal))
	}()

	// wait for container exit
	exitCh, _ := task.Wait(ctx)
	st := <-exitCh
	code, _, _ := st.Result()
	fmt.Printf("\n■ %s exited with code %d\n", name, code)
	return nil
}
