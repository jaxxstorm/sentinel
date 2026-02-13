package cli

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/jaxxstorm/sentinel/internal/constants"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newRunCmd(opts *GlobalOptions) *cobra.Command {
	var dryRun bool
	var once bool
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run Sentinel polling loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := buildRuntime(opts)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			deps.runner.Log.Info("running sentinel", zap.String("version", constants.TagName))
			return executeRun(ctx, deps, once, dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Detect diffs but do not send notifications")
	cmd.Flags().BoolVar(&once, "once", false, "Run a single poll/diff cycle and exit")
	return cmd
}

func executeRun(ctx context.Context, deps *runtimeDeps, once, dryRun bool) error {
	if once {
		err := runOnceWithTimeout(ctx, func(cctx context.Context) error {
			_, err := deps.runner.RunOnce(cctx, dryRun)
			return err
		})
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return err
	}
	err := deps.runner.Run(ctx, false, dryRun)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}
