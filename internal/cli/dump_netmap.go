package cli

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func newDumpNetmapCmd(opts *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "dump-netmap",
		Short: "Dump the current normalized netmap source payload",
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := buildRuntime(opts)
			if err != nil {
				return err
			}
			if deps.enrollment != nil {
				if _, err := deps.enrollment.EnsureEnrolled(context.Background()); err != nil {
					return err
				}
			}
			var nm any
			err = runOnceWithTimeout(context.Background(), func(ctx context.Context) error {
				polled, err := deps.source.Poll(ctx)
				if err != nil {
					return err
				}
				nm = polled
				return nil
			})
			if err != nil {
				return err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(nm)
		},
	}
}
