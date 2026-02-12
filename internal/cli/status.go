package cli

import (
	"context"
	"fmt"

	"github.com/jaxxstorm/sentinel/internal/onboarding"
	"github.com/spf13/cobra"
)

func newStatusCmd(opts *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Sentinel runtime configuration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := buildRuntime(opts)
			if err != nil {
				return err
			}
			enrollment, err := deps.enrollment.Probe(context.Background())
			if err != nil {
				// Keep status command resilient; probe errors should still render current diagnostics.
				enrollment = deps.enrollment.LastStatus()
			}
			printLine("Sentinel status")
			printLine("source_mode=%s", deps.cfg.Source.Mode)
			printLine("poll_interval=%s", deps.cfg.PollInterval)
			printLine("log_format=%s", deps.cfg.Output.LogFormat)
			printLine("detectors=%v", deps.cfg.DetectorOrder)
			printLine("state_path=%s", deps.cfg.State.Path)
			if deps.cfg.TSNet.AuthKeySource != "" && deps.cfg.TSNet.AuthKeySource != "none" {
				printLine("tailscale_auth_key_source=%s", deps.cfg.TSNet.AuthKeySource)
			}
			printEnrollmentStatus(enrollment)
			return nil
		},
	}
}

func printEnrollmentStatus(st onboarding.Status) {
	for _, line := range enrollmentStatusLines(st) {
		printLine("%s", line)
	}
}

func renderStatusSummary(st onboarding.Status) string {
	if st.State == onboarding.StateJoined {
		return fmt.Sprintf("joined mode=%s host=%s", st.Mode, st.Hostname)
	}
	if st.ErrorCode != "" {
		return fmt.Sprintf("%s code=%s", st.State, st.ErrorCode)
	}
	return string(st.State)
}

func enrollmentStatusLines(st onboarding.Status) []string {
	lines := []string{fmt.Sprintf("tailscale_status=%s", st.State)}
	if st.Mode != "" {
		lines = append(lines, fmt.Sprintf("tailscale_mode=%s", st.Mode))
	}
	if st.NodeID != "" {
		lines = append(lines, fmt.Sprintf("tailscale_node_id=%s", st.NodeID))
	}
	if st.Hostname != "" {
		lines = append(lines, fmt.Sprintf("tailscale_hostname=%s", st.Hostname))
	}
	if st.ErrorCode != "" {
		lines = append(lines, fmt.Sprintf("tailscale_error_code=%s", st.ErrorCode))
	}
	if st.Remediation != "" {
		lines = append(lines, fmt.Sprintf("tailscale_remediation=%s", st.Remediation))
	}
	if st.LoginURL != "" {
		lines = append(lines, fmt.Sprintf("tailscale_login_url=%s", st.LoginURL))
	}
	return lines
}
