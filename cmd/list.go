package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/kazuki-kanaya/codex-profiles/internal/profile"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles and remaining usage limits",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := profile.List()
		if err != nil {
			return err
		}
		writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		if _, err := fmt.Fprintln(writer, "PROFILE\tPLAN\t5H LEFT\t1W LEFT\tRESETS"); err != nil {
			return err
		}

		for _, name := range names {
			usage, err := profile.ReadUsage(name)
			if err != nil {
				cmd.PrintErrf("%s: %v\n", name, err)
			}

			plan := usage.Plan
			if plan == "" {
				plan = "-"
			}
			resets := "-"
			if usage.ResetCredits != nil {
				resets = fmt.Sprint(*usage.ResetCredits)
			}

			if _, err := fmt.Fprintf(
				writer, "%s\t%s\t%s\t%s\t%s\n",
				name, plan,
				formatRemaining(usage.FiveHourRemaining),
				formatRemaining(usage.WeeklyRemaining),
				resets,
			); err != nil {
				return err
			}
		}
		return writer.Flush()
	},
}

func formatRemaining(remaining *float64) string {
	if remaining == nil {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", *remaining)
}

func init() {
	rootCmd.AddCommand(listCmd)
}
