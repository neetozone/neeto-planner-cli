package commands

import (
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var sectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "Manage sections",
}

var sectionsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List sections in a project",
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

		data, err := c.Get(fmt.Sprintf("/projects/%s/sections", sid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects ls"},
		}
		printList(data, "sections", breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

var sectionsShowCmd = &cobra.Command{
	Use:   "show <sid>",
	Short: "View a particular section in a project",
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

		sectionSid := args[0]
		params := url.Values{}
		data, err := c.Get(fmt.Sprintf("/projects/%s/sections/%s", projectSid, sectionSid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects ls"},
			{Label: "List sections", Command: "neetoplanner sections ls"},
		}

		printResource(data, breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

var sectionsCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create a section in a project",
	Long:    "Create a section (board column) in an active project. Duplicate names are allowed. Every successful call, including a retry, creates another section.",
	Args:    cobra.ExactArgs(1),
	Example: "  $ neetoplanner sections create \"Backlog\" --project <project-sid>",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Post(fmt.Sprintf("/projects/%s/sections", projectSid), map[string]interface{}{
			"section": map[string]interface{}{"name": args[0]},
		})
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List sections", Command: fmt.Sprintf("neetoplanner sections list --project %s", projectSid)},
		}
		printActionResult(data, breadcrumbs)
		return nil
	},
}

func init() {
	addProjectFlag(sectionsListCmd)
	addPaginationFlags(sectionsListCmd)
	addVerboseFlag(sectionsListCmd)

	sectionsCmd.AddCommand(sectionsListCmd)
	sectionsListCmd.Flags().Bool("trashed", false, "List trashed sections")
	sectionsListCmd.Flags().Bool("archived", false, "List archived sections")
	sectionsListCmd.MarkFlagsMutuallyExclusive("trashed", "archived")

	addProjectFlag(sectionsShowCmd)
	addVerboseFlag(sectionsShowCmd)
	sectionsCmd.AddCommand(sectionsShowCmd)

	addProjectFlag(sectionsCreateCmd)
	sectionsCmd.AddCommand(sectionsCreateCmd)

	register(func(root *cobra.Command) { root.AddCommand(sectionsCmd) })
}
