package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

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
		results := make([]struct {
			usage profile.Usage
			err   error
		}, len(names))
		var workers sync.WaitGroup
		slots := make(chan struct{}, 4)
		for i, name := range names {
			slots <- struct{}{}
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer func() { <-slots }()
				results[i].usage, results[i].err = profile.ReadUsage(name)
			}()
		}
		workers.Wait()

		var table bytes.Buffer
		writer := tabwriter.NewWriter(&table, 0, 4, 2, ' ', 0)
		if _, err := fmt.Fprintln(writer, "PROFILE\tEMAIL\tPLAN\t5H LEFT\t1W LEFT\tRESETS"); err != nil {
			return err
		}

		for i, name := range names {
			usage := results[i].usage
			if results[i].err != nil {
				cmd.PrintErrf("%s: %v\n", name, results[i].err)
			}

			email := usage.Email
			if email == "" {
				email = "-"
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
				writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
				name, email, plan,
				formatRemaining(usage.FiveHourRemaining, usage.FiveHourResetsAt, "15:04"),
				formatRemaining(usage.WeeklyRemaining, usage.WeeklyResetsAt, "01/02 15:04"),
				resets,
			); err != nil {
				return err
			}
		}
		if err := writer.Flush(); err != nil {
			return err
		}

		header, rows, _ := strings.Cut(table.String(), "\n")
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "\x1b[1;36m%s\x1b[0m\n%s", header, rows)
		return err
	},
}

func formatRemaining(remaining *float64, resetsAt *int64, layout string) string {
	if remaining == nil {
		return "-"
	}
	value := fmt.Sprintf("%.0f%%", *remaining)
	if resetsAt != nil {
		value += " (" + time.Unix(*resetsAt, 0).Local().Format(layout) + ")"
	}
	return value
}

func init() {
	rootCmd.AddCommand(listCmd)
}
