package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

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
		printMetadata(data)
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

		listValue := args[0]
		listSid, err := resolveListSid(cmd, listValue)
		if err != nil {
			return err
		}

		params := url.Values{}
		data, err := c.Get(fmt.Sprintf("/projects/%s/lists/%s", projectSid, listSid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects"},
			{Label: "List lists", Command: "neetoplanner lists ls"},
		}

		printList(data, "todos", breadcrumbs)
		printMetadata(data)
		return nil
	},
}

func init() {
	addProjectFlag(listsListCmd)
	addPaginationFlags(listsListCmd)

	listsCmd.AddCommand(listsListCmd)
	listsListCmd.Flags().Bool("trashed", false, "List trashed lists")
	listsListCmd.Flags().Bool("archived", false, "List archived lists")
	listsListCmd.MarkFlagsMutuallyExclusive("trashed", "archived")

	addProjectFlag(listsShowCmd)
	listsCmd.AddCommand(listsShowCmd)

	rootCmd.AddCommand(listsCmd)
}

func resolveListSid(cmd *cobra.Command, listValue string) (string, error) {
	c, err := getClient(cmd)
	if err != nil {
		return "", err
	}

	projectSid, err := resolveProject(cmd)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("kind", "all")

	data, err := c.Get(fmt.Sprintf("/projects/%s/lists", projectSid), params)
	if err != nil {
		return "", err
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}

	listItems, hasItems := parsed["lists"]
	if !hasItems {
		return "", fmt.Errorf("list not found (404)")
	}

	var parsedItems []map[string]interface{}
	if err := json.Unmarshal(listItems, &parsedItems); err != nil {
		return "", err
	}
	for _, elem := range parsedItems {
		sid, _ := elem["sid"].(string)
		if strings.EqualFold(sid, listValue) {
			return sid, nil
		}
	}
	trimmedValue := strings.TrimSpace(listValue)
	for _, elem := range parsedItems {
		name, _ := elem["name"].(string)
		if strings.EqualFold(strings.TrimSpace(name), trimmedValue) {
			return elem["sid"].(string), nil
		}
	}
	return "", fmt.Errorf("no list matching %q found", listValue)
}
