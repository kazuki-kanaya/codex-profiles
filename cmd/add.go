package cmd

import (
	"github.com/kazuki-kanaya/codex-profiles/internal/profile"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add NAME",
	Short: "Create a profile and log in to Codex",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := validateProfileName(cmd, args[0]); err != nil {
			return err
		}
		return profile.Add(name)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
