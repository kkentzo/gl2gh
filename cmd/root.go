package cmd

import "github.com/spf13/cobra"

type GlobalVariables struct {
	ExportPath   string
	UserMappings map[string]int
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

	return root
}

func requireGlobalFlags(cmd *cobra.Command, globals *GlobalVariables) *cobra.Command {
	cmd.Flags().StringVarP(&globals.ExportPath, "export", "e", "", "directory that contains the uncompressed gitlab export")
	cmd.Flags().StringToIntVarP(&globals.UserMappings, "users", "u", map[string]int{}, "mapping of github user names to gitlab UIDs")
	cmd.MarkFlagRequired("export")
	return cmd
}
