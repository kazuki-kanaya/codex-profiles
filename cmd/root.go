package cmd

import (
	"fmt"
	"os"

	"github.com/kazuki-kanaya/codex-profiles/internal/profile"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "codex-profiles",
	Short: "Manage separate Codex account profiles",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func validateProfileName(cmd *cobra.Command, name string) error {
	if err := profile.ValidateName(name); err != nil {
		return err
	}
	for _, command := range cmd.Root().Commands() {
		if name == command.Name() || command.HasAlias(name) {
			return fmt.Errorf("profile name %q is reserved for the %q command", name, command.Name())
		}
	}
	return nil
}
