package cmd

import (
	"bufio"
	"strings"

	"github.com/kazuki-kanaya/codex-profiles/internal/profile"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename OLD NEW",
	Short: "Rename a profile with a fresh Codex environment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName, newName := args[0], args[1]
		if err := validateProfileName(cmd, newName); err != nil {
			return err
		}
		if err := profile.ValidateRename(oldName, newName); err != nil {
			return err
		}

		cmd.Printf("Close Codex for profile %q before continuing. Its daemon will be stopped automatically.\n", oldName)
		cmd.Printf("Rename profile %q to %q? [y/N]: ", oldName, newName)
		scanner := bufio.NewScanner(cmd.InOrStdin())
		if !scanner.Scan() {
			return scanner.Err()
		}
		if !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
			return nil
		}
		return profile.Rename(oldName, newName)
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
