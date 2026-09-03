package commands

import (
	"github.com/neetozone/neeto-planner-cli/internal/output"
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
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		kind := kindParam(cmd, "template", "archived", "trashed", "all")
		params := paginationParams(cmd)
		if kind != "" {
			params.Set("kind", kind)
		}

		data, err := c.Get("/projects", params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List lists in a project", Command: "neetoplanner lists ls --project <sid>"},
		}

		printList(data, "projects", breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

func init() {
	addPaginationFlags(projectsListCmd)
	addVerboseFlag(projectsListCmd)

	projectsCmd.AddCommand(projectsListCmd)
	projectsListCmd.Flags().Bool("template", false, "List template projects")
	projectsListCmd.Flags().Bool("archived", false, "List archived projects")
	projectsListCmd.Flags().Bool("trashed", false, "List trashed projects")
	projectsListCmd.Flags().Bool("all", false, "List all projects regardless of kind")
	projectsListCmd.MarkFlagsMutuallyExclusive("template", "archived", "trashed", "all")

	rootCmd.AddCommand(projectsCmd)
}
