package cli

import (
	"fmt"
	"io"

	"github.com/jaxxstorm/sentinel/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd(_ *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show build and version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := version.Current()
			writeVersion(cmd.OutOrStdout(), info)
			return nil
		},
	}
}

func writeVersion(w io.Writer, info version.Metadata) {
	fmt.Fprintf(w, "Version: %s\n", info.Version)
	fmt.Fprintf(w, "Build Timestamp: %s\n", info.BuildTimestamp)
	fmt.Fprintf(w, "Commit Hash: %s\n", info.CommitHash)
}
