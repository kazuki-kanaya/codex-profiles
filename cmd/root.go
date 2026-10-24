package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/kazuki-kanaya/codex-profiles/internal/profile"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                "codex-profiles [PROFILE] [codex args...]",
	Short:              "Manage separate Codex account profiles",
	DisableFlagParsing: true,
	Args:               cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 ||
			(len(args) == 1 && (args[0] == "-h" || args[0] == "--help")) {
			return cmd.Help()
		}
		if len(args) == 1 && args[0] == "--version" {
			return printVersion(cmd, args)
		}

		name := args[0]
		if err := validateProfileName(cmd, name); err != nil {
			return err
		}
		return profile.Run(name, args[1:]...)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 {
			os.Exit(exitErr.ExitCode())
		}
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
