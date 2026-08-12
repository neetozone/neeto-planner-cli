package commands

import (
	"github.com/spf13/cobra"
)

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage lists",
}

var listsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List lists in a project",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := resolveProject(cmd); err != nil {
			return err
		}
		return notImplemented("GET /projects/:project_id/lists")
	},
}

func init() {
	addProjectFlag(listsListCmd)
	addPaginationFlags(listsListCmd)

	listsCmd.AddCommand(listsListCmd)

	rootCmd.AddCommand(listsCmd)
}
