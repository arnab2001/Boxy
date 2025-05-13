package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/arnab2001/boxy/internal/client"
	"github.com/containerd/containerd"
	refdocker "github.com/containerd/containerd/reference/docker"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "pull <image>",
		Short: "Pull an OCI image into containerd (shows spinner)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			named, err := refdocker.ParseDockerRef(args[0])
			if err != nil {
				return err
			}
			ref := named.String()

			ctx, cancel := context.WithCancel(client.Default())
			defer cancel()

			c, err := client.Instance()
			if err != nil {
				return err
			}

			// ── very small spinner ──────────────────────────────
			done := make(chan struct{})
			go func() {
				chars := []rune{'|', '/', '-', '\\'}
				i := 0
				for {
					select {
					case <-done:
						return
					default:
						fmt.Fprintf(os.Stdout, "\r%c pulling %s …", chars[i%4], ref)
						i++
						time.Sleep(150 * time.Millisecond)
					}
				}
			}()

			_, err = c.Pull(ctx, ref, containerd.WithPullUnpack)
			close(done)
			if err != nil {
				fmt.Fprintf(os.Stdout, "\r✖ failed to pull %s: %v\n", ref, err)
				return err
			}
			fmt.Fprintf(os.Stdout, "\r✔ pulled %s\n", ref)
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
