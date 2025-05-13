package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/arnab2001/boxy/internal/client"
	"github.com/containerd/containerd/errdefs"
	"github.com/spf13/cobra"
)

const defaultTimeout = 10 * time.Second // like docker

func init() {
	cmd := &cobra.Command{
		Use:   "stop <name> [timeout]",
		Short: "Gracefully stop a running container",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(_ *cobra.Command, args []string) error {
			timeout := defaultTimeout
			if len(args) == 2 {
				if t, err := time.ParseDuration(args[1]); err == nil {
					timeout = t
				}
			}

			ctx := client.Default()
			c, err := client.Instance()
			if err != nil {
				return err
			}

			cont, err := c.LoadContainer(ctx, args[0])
			if err != nil {
				return err
			}

			taskObj, err := cont.Task(ctx, nil)
			if errdefs.IsNotFound(err) {
				fmt.Printf("✓ %s already stopped\n", args[0])
				return nil
			}
			if err != nil {
				return err
			}

			// 1) try SIGTERM
			if err := taskObj.Kill(ctx, syscall.SIGTERM); err != nil &&
				!errdefs.IsNotFound(err) {
				return err
			}

			exitCh, _ := taskObj.Wait(ctx)

			select {
			case <-exitCh:
				fmt.Printf("✓ stopped %s\n", args[0])
				return nil
			case <-time.After(timeout):
				// 2) escalate to SIGKILL
				if err := taskObj.Kill(ctx, syscall.SIGKILL); err != nil &&
					!errdefs.IsNotFound(err) {
					return err
				}
				<-exitCh
				fmt.Printf("✓ force-stopped %s\n", args[0])
				return nil
			}
		},
	}
	rootCmd.AddCommand(cmd)
}
