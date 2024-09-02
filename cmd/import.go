package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/kkentzo/gl-to-gh/github"
	"github.com/kkentzo/gl-to-gh/gitlab"
	"github.com/spf13/cobra"
)

func ImportCommand(globals *GlobalVariables) *cobra.Command {
	var (
		commentsOnly bool
		reverse      bool
		delay        time.Duration
		startFromId  int
		repo         string
		token        string
		labels       []string
		dryRun       bool

		descr = "Mass import of gitlab issues to github"
		cmd   = &cobra.Command{
			Use:   "import",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				client := github.NewClient(token, dryRun, globals.Debug)

				t0 := time.Now()
				log.Printf("Starting from ID=%d [delay=%v]", startFromId, delay)
				defer func() { log.Printf("Duration: %v\nRequest Count: %d", time.Now().Sub(t0), client.RequestCount()) }()

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
				start := startFromId
				end := last.Id
				if commentsOnly {
					if reverse {
						start, end = end, start
						log.Printf("Reversing order [comments=%v] [reverse=%v] [start=%d] [end=%d]", commentsOnly, reverse, start, end)
					} else {
						log.Printf("--reverse can be specified only in conjuction with --comments [reverse=%v] [comments=%v]",
							reverse, commentsOnly)
						return
					}
				}

				iid := start
				for {
					// check iteration
					if reverse {
						if iid < end {
							break
						}
					} else {
						if iid > end {
							break
						}
					}

					// perform operations
					if commentsOnly {
						if issue, err := PostComments(iid, issueMap, client, repo, labels, delay); err != nil {
							log.Printf("%v", err)
							return
						} else {
							if issue == nil {
								log.Printf("[#%d] Issue does not exist (0 comments)", iid)
							} else {
								log.Printf("[#%d] %s (%d comments)", iid, issue.Title, len(issue.Comments()))
							}
						}
					} else {
						if issue, err := PostIssue(iid, issueMap, client, repo, labels); err != nil {
							log.Printf("%v", err)
							return
						} else {
							log.Printf("[#%d] %s (%d comments)", iid, issue.Title, len(issue.Comments()))
						}
						time.Sleep(delay)
					}

					// advance iteration
					if reverse {
						iid--
					} else {
						iid++
					}
				}
			},
		}
	)

	cmd.Flags().BoolVar(&commentsOnly, "comments", false, "import comments only (assumes all issues have been imported and have the same IDs)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "reverse the order of issue IDs (only for comments)")
	cmd.Flags().IntVar(&startFromId, "start", 1, "ID to start the migration from (lower IDs will be skipped)")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "the target github repo in the form 'user_or_org/repo_name'")
	cmd.Flags().StringVarP(&token, "token", "t", "", "the API token for authenticating with github API")
	cmd.Flags().DurationVar(&delay, "delay", time.Duration(10*time.Second), "delay between successive API calls")
	cmd.Flags().StringSliceVarP(&labels, "labels", "l", []string{}, "a comma-separated list of labels to be attached to the issue")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if true then no API call will be made")
	cmd.MarkFlagRequired("repo")
	cmd.MarkFlagRequired("token")
	return requireGlobalFlags(cmd, globals, []string{"export"})
}

func PostIssue(iid int, issueMap map[int]*github.Issue, client *github.Client, repo string, labels []string) (*github.Issue, error) {
	if issue, ok := issueMap[iid]; ok {
		if err := issue.Post(client, repo); err != nil {
			return issue, fmt.Errorf("[#%d] failed to POST issue: %v\n", iid, err)
		} else {
			return issue, nil
		}
	} else {
		// create placeholder issue
		issue := github.NewPlaceholder(labels)
		if err := issue.Post(client, repo); err != nil {
			return issue, fmt.Errorf("[#%d] failed to POST placeholder issue: %v\n", iid, err)
		} else {
			return issue, nil
		}
	}
}

func PostComments(iid int, issueMap map[int]*github.Issue, client *github.Client, repo string, labels []string, delay time.Duration) (*github.Issue, error) {
	if issue, ok := issueMap[iid]; ok {
		for _, comment := range issue.Comments() {
			if err := comment.Post(client, repo, iid); err != nil {
				return nil, fmt.Errorf("[#%d] failed to post comment: %v", iid, err)
			}
			time.Sleep(delay)
		}
		return issue, nil
	}
	return nil, nil
}
