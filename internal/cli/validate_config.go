package cli

import (
	"github.com/jaxxstorm/sentinel/internal/config"
	"github.com/spf13/cobra"
)

func newValidateConfigCmd(opts *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "validate-config",
		Short: "Validate Sentinel configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(opts.ConfigPath)
			if err != nil {
				return err
			}
			if err := config.Validate(cfg); err != nil {
				return err
			}
			printLine("configuration valid")
			return nil
		},
	}
}
