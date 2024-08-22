package cmd

import (
	"fmt"

	"github.com/kkentzo/gl-to-gh/gitlab"
	"github.com/spf13/cobra"
)

func ShowCommand(globals *GlobalVariables) *cobra.Command {
	var (
		issueId uint

		descr = "Display a specific issue with its comments"
		cmd   = &cobra.Command{
			Use:   "show",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				// parse ndjson
				issues, err := gitlab.Parse(globals.ExportPath)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "error: %v\n", err)
					return
				}

				// find issue
				var issue *gitlab.Issue
				for _, iss := range issues {
					if iss.Id == int(issueId) {
						issue = iss
						break
					}
				}
				if issue == nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Issue with id=%d was not found in exprt file\n", issueId)
					return
				}

				fmt.Println(issue.Summarize())

				fmt.Println(issue.Convert(map[int]string{}))

				for _, comment := range issue.Comments {
					fmt.Println(comment.Convert(map[int]string{}))
					fmt.Println("=============================================")
				}
			},
		}
	)

	cmd.Flags().UintVar(&issueId, "id", 0, "the ID of the issue to be displated")
	cmd.MarkFlagRequired("id")
	return requireGlobalFlags(cmd, globals)
}
