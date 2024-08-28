package cmd

import (
	"fmt"

	"github.com/kkentzo/gl-to-gh/gitlab"
	"github.com/spf13/cobra"
)

func UsersCommand(globals *GlobalVariables) *cobra.Command {
	var (
		descr = "Display the unique gitlab user IDs found across issues and comments"
		cmd   = &cobra.Command{
			Use:   "users",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				// parse ndjson
				issues, err := gitlab.Parse(globals.ExportPath, globals.CommentExclusionFilter)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "error: %v\n", err)
					return
				}

				// find unique users in issues only
				uids := map[int]bool{}
				for _, issue := range issues {
					uids[issue.AuthorId] = true
				}

				fmt.Fprintln(cmd.OutOrStderr(), "Unique User IDs in Issues:")

				for uid, _ := range uids {
					fmt.Fprintf(cmd.OutOrStderr(), "%d\n", uid)
				}
			},
		}
	)

	return requireGlobalFlags(cmd, globals, []string{"export"})
}
