package cmd

import (
	"fmt"

	"github.com/kkentzo/gl-to-gh/gitlab"
	"github.com/spf13/cobra"
)

func SummaryCommand(globals *GlobalVariables) *cobra.Command {
	var (
		descr = "Display a list of all the discovered issues"
		cmd   = &cobra.Command{
			Use:   "summary",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				// parse ndjson
				issues, err := gitlab.Parse(globals.ExportPath)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "error: %v\n", err)
				}

				nc := 0

				for _, issue := range issues {
					fmt.Fprintf(cmd.OutOrStderr(), issue.Summarize())
					nc += len(issue.Comments)
				}

				fmt.Fprintf(cmd.OutOrStderr(), "Issues: %d\n", len(issues))
				fmt.Fprintf(cmd.OutOrStderr(), "Comments: %d\n", nc)
			},
		}
	)

	return requireGlobalFlags(cmd, globals)
}
