package cmd

import (
	"fmt"
	"time"

	"github.com/kkentzo/gl-to-gh/github"
	"github.com/spf13/cobra"
)

func RateCommand(globals *GlobalVariables) *cobra.Command {
	var (
		token string

		descr = "query the rate limits of github's API"
		cmd   = &cobra.Command{
			Use:   "rate",
			Short: descr,
			Long:  descr,
			Run: func(cmd *cobra.Command, args []string) {
				client := github.NewClient(token, false, globals.Debug)
				rate, err := client.RateLimit()
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "error: %v\n", err)
					return
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Retrieved at: %v\n", time.Now())
				fmt.Fprintf(cmd.OutOrStdout(), "Used: %d/%d\n", rate.Used, rate.Limit)
				fmt.Fprintf(cmd.OutOrStdout(), "Remaining: %d\n", rate.Remaining)
				fmt.Fprintf(cmd.OutOrStdout(), "ResetAt: %v\n", rate.ResetAt)
			},
		}
	)

	cmd.Flags().StringVarP(&token, "token", "t", "", "the API token for authenticating with github API")
	cmd.MarkFlagRequired("token")
	return requireGlobalFlags(cmd, globals, []string{})
}
