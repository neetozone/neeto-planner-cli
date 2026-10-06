package commands

import (
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
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

		kind := kindParam(cmd, "trashed", "archived")
		params := paginationParams(cmd)
		params.Add("kind", kind)

		data, err := c.Get(fmt.Sprintf("/projects/%s/lists", sid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects ls"},
		}
		printList(data, "lists", breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

var listsShowCmd = &cobra.Command{
	Use:   "show <sid>",
	Short: "View a particular list in a project",
	Args:  cobra.ExactArgs(1),
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
			{Label: "List projects", Command: "neetoplanner projects ls"},
			{Label: "List lists", Command: "neetoplanner lists ls"},
		}

		printResource(data, breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

var listsCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create a list in a project",
	Long:    "Create a list (board column) in an active project. Duplicate names are allowed. Every successful call, including a retry, creates another list.",
	Args:    cobra.ExactArgs(1),
	Example: "  $ neetoplanner lists create \"Backlog\" --project <project-sid>",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Post(fmt.Sprintf("/projects/%s/lists", projectSid), map[string]interface{}{
			"list": map[string]interface{}{"name": args[0]},
		})
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List lists", Command: fmt.Sprintf("neetoplanner lists list --project %s", projectSid)},
		}
		printActionResult(data, breadcrumbs)
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

	addProjectFlag(listsCreateCmd)
	listsCmd.AddCommand(listsCreateCmd)

	register(func(root *cobra.Command) { root.AddCommand(listsCmd) })
}
