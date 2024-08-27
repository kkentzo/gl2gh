package cmd

import "github.com/spf13/cobra"

type GlobalVariables struct {
	ExportPath             string
	CommentExclusionFilter []string
	UserMappings           map[string]int
	ReplacePatterns        map[string]string
	Debug                  bool
}

var DefaultCommentExclusionFilter = []string{
	"mentioned in",
	"assigned to",
	"unassigned",
	"changed the description",
	"created branch",
	"changed title",
	"marked the checklist",
	"marked this issue",
}

func ReverseMapping(mapping map[string]int) map[int]string {
	reverse := map[int]string{}
	for username, uid := range mapping {
		reverse[uid] = username
	}
	return reverse
}

func New() *cobra.Command {
	globals := &GlobalVariables{}

	descr := "Migrate issues from gitlab to github"

	root := &cobra.Command{
		Use:   "gl2gh",
		Short: descr,
		Long:  descr,
	}

	root.AddCommand(SummaryCommand(globals))
	root.AddCommand(ShowCommand(globals))
	root.AddCommand(UsersCommand(globals))
	root.AddCommand(PostCommand(globals))
	root.AddCommand(ImportCommand(globals))

	return root
}

func requireGlobalFlags(cmd *cobra.Command, globals *GlobalVariables) *cobra.Command {
	cmd.Flags().StringVarP(&globals.ExportPath, "export", "e", "", "directory that contains the uncompressed gitlab export")
	cmd.Flags().StringSliceVarP(&globals.CommentExclusionFilter, "filter", "f", DefaultCommentExclusionFilter, "exclude comments that start with the supplied substrings")
	cmd.Flags().StringToIntVarP(&globals.UserMappings, "users", "u", map[string]int{}, "mapping of github user names to gitlab UIDs")
	cmd.Flags().StringToStringVar(&globals.ReplacePatterns, "replace", map[string]string{},
		"specify pairs of replacement patterns for issue and comment texts (useful for replacing link URIs)")
	cmd.Flags().BoolVarP(&globals.Debug, "debug", "d", false, "whether to display debugging information")
	cmd.MarkFlagRequired("export")
	return cmd
}
