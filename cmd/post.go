package cmd

import (
	"fmt"

	"github.com/kkentzo/gl-to-gh/github"
	"github.com/kkentzo/gl-to-gh/gitlab"
	"github.com/spf13/cobra"
)

func PostCommand(globals *GlobalVariables) *cobra.Command {
	var (
		issueId uint
		repo    string
		token   string
		labels  []string
		dryRun  bool

		descr = "Post a specific issue and its comments to a gitlab repo"
		cmd   = &cobra.Command{
			Use:   "post",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				mappings := ReverseMapping(globals.UserMappings)

				// parse ndjson
				issues, err := gitlab.Parse(globals.ExportPath)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "error: %v\n", err)
					return
				}

				// find the issue
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

				// ok, let's now post the issue
				client := github.NewClient(token, dryRun, globals.Debug)
				ghIssue := github.New(issue, mappings, labels)
				if err := ghIssue.Post(client, repo); err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Posting error: %v\n", err)
				}
			},
		}
	)

	cmd.Flags().UintVar(&issueId, "id", 0, "the ID of the issue to be displated")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "the target github repo in the form 'user_or_org/repo_name'")
	cmd.Flags().StringVarP(&token, "token", "t", "", "the API token for authenticating with github API")
	cmd.Flags().StringSliceVarP(&labels, "labels", "l", []string{}, "a comma-separated list of labels to be attached to the issue")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if true then no API call will be made")
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("repo")
	cmd.MarkFlagRequired("token")
	return requireGlobalFlags(cmd, globals)
}
