package cmd

import (
	"log"
	"time"

	"github.com/kkentzo/gl-to-gh/github"
	"github.com/kkentzo/gl-to-gh/gitlab"
	"github.com/spf13/cobra"
)

func ImportCommand(globals *GlobalVariables) *cobra.Command {
	var (
		delay  time.Duration
		repo   string
		token  string
		labels []string
		dryRun bool

		descr = "Mass import of gitlab issues to github"
		cmd   = &cobra.Command{
			Use:   "import",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				client := github.NewClient(token, dryRun, globals.Debug)

				t0 := time.Now()
				log.Printf("Starting with delay=%v", delay)
				defer log.Printf("Duration: %v\nRequest Count: %d", time.Now().Sub(t0), client.RequestCount())

				mappings := ReverseMapping(globals.UserMappings)

				// parse ndjson
				issues, err := gitlab.Parse(globals.ExportPath, globals.CommentExclusionFilter)
				if err != nil {
					log.Printf("error: %v", err)
					return
				}

				if len(issues) == 0 {
					log.Printf("no issues found in file %s", globals.ExportPath)
					return
				}

				issueMap := map[int]*github.Issue{}

				for _, issue := range issues {
					issueMap[issue.Id], err = github.New(issue, mappings, labels, globals.ReplacePatterns)
					if err != nil {
						log.Printf("[#%d] failed to convert issue: %v", issue.Id, err)
						return
					}
				}

				last := issues[len(issues)-1]

				for i := 1; i <= last.Id; i++ {
					if issue, ok := issueMap[i]; ok {
						if err = issue.Post(client, repo, delay); err != nil {
							log.Printf("[#%d] failed to POST issue: %v\n", i, err)
							return
						} else {
							log.Printf("[#%d] %s (%d comments)", i, issue.Title, len(issue.Comments()))
						}
					} else {
						// create placeholder issue
						if err = github.NewPlaceholder(labels).Post(client, repo, delay); err != nil {
							log.Printf("[#%d] failed to POST placeholder issue: %v\n", i, err)
							return
						} else {
							log.Printf("[#%d] PLACEHOLDER", i)
						}
					}
					time.Sleep(delay)
				}
			},
		}
	)

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "the target github repo in the form 'user_or_org/repo_name'")
	cmd.Flags().StringVarP(&token, "token", "t", "", "the API token for authenticating with github API")
	cmd.Flags().DurationVar(&delay, "delay", time.Duration(10*time.Second), "delay between successive API calls")
	cmd.Flags().StringSliceVarP(&labels, "labels", "l", []string{}, "a comma-separated list of labels to be attached to the issue")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if true then no API call will be made")
	cmd.MarkFlagRequired("repo")
	cmd.MarkFlagRequired("token")
	return requireGlobalFlags(cmd, globals, []string{"export"})
}
