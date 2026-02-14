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
			if deps.cfg.TSNet.CredentialMode != "" && deps.cfg.TSNet.CredentialMode != "none" {
				printLine("tailscale_credential_mode=%s", deps.cfg.TSNet.CredentialMode)
			}
			if deps.cfg.TSNet.CredentialSource != "" && deps.cfg.TSNet.CredentialSource != "none" {
				printLine("tailscale_credential_source=%s", deps.cfg.TSNet.CredentialSource)
			}
			if len(deps.cfg.TSNet.AdvertiseTags) > 0 {
				printLine("tailscale_advertise_tags=%v", deps.cfg.TSNet.AdvertiseTags)
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
	if st.CredentialSource != "" && st.CredentialSource != "none" {
		lines = append(lines, fmt.Sprintf("tailscale_credential_source=%s", st.CredentialSource))
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
