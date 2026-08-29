package commands

import (
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-planner-cli/internal/output"
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
		sid, err := resolveProject(cmd)
		if err != nil {
			return err
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		var kind string
		switch {
		case cmd.Flags().Changed("trashed"):
			kind = "trashed"
		case cmd.Flags().Changed("archived"):
			kind = "archived"
		default:
			kind = ""
		}

		params := url.Values{}
		params.Add("kind", kind)

		data, err := c.Get(fmt.Sprintf("/projects/%s/lists", sid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects"},
		}
		printList(data, "lists", breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

var listsShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "View a particular list in a project",
	Aliases: []string{"lsh"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		listSid := args[0]
		params := url.Values{}
		data, err := c.Get(fmt.Sprintf("/projects/%s/lists/%s", projectSid, listSid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects"},
			{Label: "List lists", Command: "neetoplanner lists ls"},
		}

		printResource(data, breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

func init() {
	addProjectFlag(listsListCmd)
	addPaginationFlags(listsListCmd)
	addVerboseFlag(listsListCmd)

	listsCmd.AddCommand(listsListCmd)
	listsListCmd.Flags().Bool("trashed", false, "List trashed lists")
	listsListCmd.Flags().Bool("archived", false, "List archived lists")
	listsListCmd.MarkFlagsMutuallyExclusive("trashed", "archived")

	addProjectFlag(listsShowCmd)
	addVerboseFlag(listsShowCmd)
	listsCmd.AddCommand(listsShowCmd)

	rootCmd.AddCommand(listsCmd)
}
