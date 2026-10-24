package cmd

import (
	"fmt"

	"github.com/kazuki-kanaya/codex-profiles/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Args:  cobra.NoArgs,
	RunE:  printVersion,
}

func printVersion(cmd *cobra.Command, _ []string) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", cmd.Root().Name(), version.Version)
	return err
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
