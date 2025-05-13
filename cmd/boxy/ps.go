package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/arnab2001/boxy/internal/client"
	"github.com/containerd/containerd/errdefs"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List containers in the boxy namespace",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := client.Default()
			c, err := client.Instance()
			if err != nil {
				return err
			}

			containers, err := c.Containers(ctx)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tSTATE\tPID\tIMAGE")

			for _, cont := range containers {
				info, _ := cont.Info(ctx)

				state := "UNKNOWN"
				pid := "-"

				taskObj, err := cont.Task(ctx, nil)
				switch {
				case err == nil:
					st, serr := taskObj.Status(ctx)
					if serr == nil {
						state = string(st.Status) // RUNNING / STOPPED / etc.
					}
					if p := taskObj.Pid(); p != 0 {
						pid = fmt.Sprint(p)
					}
				case errdefs.IsNotFound(err):
					// container exists but no running task
					state = "STOPPED"
				default:
					return err // real error
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					info.ID, state, pid, info.Image)
			}
			return w.Flush()
		},
	}
	rootCmd.AddCommand(cmd)
}
