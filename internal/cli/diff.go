package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newDiffCmd(opts *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Run one diff cycle and print formatted results",
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := buildRuntime(opts)
			if err != nil {
				return err
			}
			var resEvents string
			err = runOnceWithTimeout(context.Background(), func(ctx context.Context) error {
				res, err := deps.runner.RunOnce(ctx, true)
				if err != nil {
					return err
				}
				resEvents = deps.renderer.FormatDiff(res.Events)
				return nil
			})
			if err != nil {
				return err
			}
			printLine(resEvents)
			return nil
		},
	}
}
