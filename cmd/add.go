package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add NAME",
	Short: "Create a profile and log in to Codex",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateProfileName(cmd, args[0]); err != nil {
			return err
		}
		fmt.Println("add called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
