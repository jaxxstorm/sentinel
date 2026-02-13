package cli

import (
	"fmt"

	"github.com/jaxxstorm/sentinel/internal/constants"
	"github.com/spf13/cobra"
)

func newVersionCmd(_ *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show build and version information",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("Version: %s\n", constants.TagName)
			fmt.Printf("Build Timestamp: %s\n", constants.BuildTimestamp)
			fmt.Printf("Commit Hash: %s\n", constants.CommitHash)
			return nil
		},
	}
}
