package cmd

import "github.com/spf13/cobra"

type GlobalVariables struct {
	ExportPath string
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

	return root
}

func requireGlobalFlags(cmd *cobra.Command, globals *GlobalVariables) *cobra.Command {
	cmd.Flags().StringVarP(&globals.ExportPath, "export", "e", "", "directory that contains the uncompressed gitlab export")
	cmd.MarkFlagRequired("export")
	return cmd
}
