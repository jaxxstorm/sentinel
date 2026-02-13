package cli

import (
	"time"

	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	ConfigPath                string
	LogFormat                 string
	LogLevel                  string
	NoColor                   bool
	TailscaleAuthKey          string
	TailscaleLoginMode        string
	TailscaleStateDir         string
	TailscaleLoginTimeout     time.Duration
	TailscaleFallbackOverride bool
}

func NewRootCommand() *cobra.Command {
	opts := &GlobalOptions{}
	cmd := &cobra.Command{
		Use:   "sentinel",
		Short: "Tailnet netmap diff and notification daemon",
	}
	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "Path to config file (yaml or json)")
	cmd.PersistentFlags().StringVar(&opts.LogFormat, "log-format", "", "Log format: pretty|json")
	cmd.PersistentFlags().StringVar(&opts.LogLevel, "log-level", "", "Log level")
	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable ANSI color output")
	cmd.PersistentFlags().StringVar(&opts.TailscaleAuthKey, "tailscale-auth-key", "", "Tailscale auth key for node onboarding")
	cmd.PersistentFlags().StringVar(&opts.TailscaleLoginMode, "tailscale-login-mode", "", "Tailscale onboarding mode: auto|auth_key|interactive")
	cmd.PersistentFlags().StringVar(&opts.TailscaleStateDir, "tailscale-state-dir", "", "Tailscale tsnet state directory")
	cmd.PersistentFlags().DurationVar(&opts.TailscaleLoginTimeout, "tailscale-login-timeout", 0, "Timeout for interactive tailscale login")
	cmd.PersistentFlags().BoolVar(&opts.TailscaleFallbackOverride, "tailscale-allow-interactive-fallback", false, "Allow fallback to interactive login after auth key failure")

	cmd.AddCommand(newRunCmd(opts))
	cmd.AddCommand(newStatusCmd(opts))
	cmd.AddCommand(newDiffCmd(opts))
	cmd.AddCommand(newDumpNetmapCmd(opts))
	cmd.AddCommand(newTestNotifyCmd(opts))
	cmd.AddCommand(newValidateConfigCmd(opts))
	cmd.AddCommand(newVersionCmd(opts))
	return cmd
}
