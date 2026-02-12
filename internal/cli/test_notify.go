package cli

import (
	"context"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/spf13/cobra"
)

func newTestNotifyCmd(opts *GlobalOptions) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "test-notify",
		Short: "Send a synthetic test notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := buildRuntime(opts)
			if err != nil {
				return err
			}
			evt := event.NewPresenceEvent(event.TypePeerOnline, "test-peer", "before", "after", map[string]any{"name": "test-peer"}, time.Now())
			result, err := deps.notifier.Notify(context.Background(), []event.Event{evt}, dryRun)
			if err != nil {
				return err
			}
			printLine("notifications sent=%d dry_run=%d suppressed=%d", result.Sent, result.DryRun, result.Suppressed)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Do not send outbound sink requests")
	return cmd
}
