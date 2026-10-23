package cmd

import (
	"bufio"
	"strings"

	"github.com/kazuki-kanaya/codex-profiles/internal/profile"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove NAME",
	Short: "Remove a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := profile.ValidateName(name); err != nil {
			return err
		}

		cmd.Printf("Remove profile %q? [y/N]: ", name)

		scanner := bufio.NewScanner(cmd.InOrStdin())
		if !scanner.Scan() {
			return scanner.Err()
		}

		answer := strings.TrimSpace(scanner.Text())
		if !strings.EqualFold(answer, "y") {
			return nil
		}

		return profile.Remove(name)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
