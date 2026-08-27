package commands

import (
	"encoding/json"
	"fmt"
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
		case getBoolFlag(cmd, "template"):
			kind = "template"
		case getBoolFlag(cmd, "archived"):
			kind = "archived"
		case getBoolFlag(cmd, "trashed"):
			kind = "trashed"
		case getBoolFlag(cmd, "all"):
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

		fmt.Printf("\nOrganization: %s\n\n", metadata["organization"])
		printList(data, "projects", nil)
		if count, ok := metadata["total_count"].(float64); ok {
			fmt.Printf("\nTotal count: %d\n", int(count))
		}
		return nil
	},
}

func init() {
	addPaginationFlags(projectsListCmd)

	projectsCmd.AddCommand(projectsListCmd)
	projectsListCmd.Flags().Bool("template", false, "List template projects")
	projectsListCmd.Flags().Bool("archived", false, "List archived projects")
	projectsListCmd.Flags().Bool("trashed", false, "List trashed projects")
	projectsListCmd.Flags().Bool("all", false, "List all projects regardless of kind")
	projectsListCmd.MarkFlagsMutuallyExclusive("template", "archived", "trashed", "all")

	rootCmd.AddCommand(projectsCmd)
}
