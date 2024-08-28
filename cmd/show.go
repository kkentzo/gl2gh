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
				mappings := ReverseMapping(globals.UserMappings)

				// parse ndjson
				issues, err := gitlab.Parse(globals.ExportPath, globals.CommentExclusionFilter)
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
					fmt.Fprintf(cmd.OutOrStderr(), "Issue with id=%d was not found in export file\n", issueId)
					return
				}

				fmt.Println(issue.Summarize())

				s, err := issue.Convert(mappings, globals.ReplacePatterns)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Error converting issue %d: %v", issueId, err)
					return
				}
				fmt.Println(s)

				for _, comment := range issue.Comments {
					s, err := comment.Convert(globals.ReplacePatterns)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStderr(), "Error converting comment for issue %d: %v", issueId, err)
						return
					}

					fmt.Println(s)
					fmt.Println("=============================================")
				}
			},
		}
	)

	cmd.Flags().UintVar(&issueId, "id", 0, "the ID of the issue to be displated")
	cmd.MarkFlagRequired("id")
	return requireGlobalFlags(cmd, globals, []string{"export"})
}
