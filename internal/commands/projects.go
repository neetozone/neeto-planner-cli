package commands

import (
	"encoding/json"
	"net/url"

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

		var kind string
		switch {
		case cmd.Flags().Changed("template"):
			kind = "template"
		case cmd.Flags().Changed("archived"):
			kind = "archived"
		case cmd.Flags().Changed("trashed"):
			kind = "trashed"
		case cmd.Flags().Changed("all"):
			kind = "all"
		default:
			kind = ""
		}

		params := url.Values{}
		params.Add("kind", kind)

		data, err := c.Get("/projects", params)
		if err != nil {
			return err
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return err
		}

		printList(data, "projects", nil)
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
