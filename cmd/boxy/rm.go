package main

import (
	"fmt"
	"syscall"

	"github.com/arnab2001/boxy/internal/client"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Kill & remove a container and its snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
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
			if err == nil {
				// 1) ask the task to die
				if err := taskObj.Kill(ctx, syscall.SIGKILL); err != nil &&
					!errdefs.IsNotFound(err) {
					return err
				}

				// 2) wait until shim reports exit, idk how shim actually works
				exitCh, _ := taskObj.Wait(ctx)
				<-exitCh

				// 3) delete task / shim
				if _, err := taskObj.Delete(ctx); err != nil &&
					!errdefs.IsNotFound(err) {
					return err
				}
			} else if !errdefs.IsNotFound(err) {
				return err // i think this will look up error
			}

			if err := cont.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
				return err
			}
			fmt.Printf("✓ removed %s\n", args[0])
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
