package commands

import (
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List projects",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return notImplemented("GET /projects")
	},
}

func init() {
	addPaginationFlags(projectsListCmd)

	projectsCmd.AddCommand(projectsListCmd)

	rootCmd.AddCommand(projectsCmd)
}
