package commands

import (
	"fmt"
	"net/url"

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
		case getBoolFlag(cmd, "trashed"):
			kind = "trashed"
		case getBoolFlag(cmd, "archived"):
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

		printList(data, "lists", nil)
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

	rootCmd.AddCommand(listsCmd)
}
