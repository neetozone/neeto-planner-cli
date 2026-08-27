package commands

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var todosCmd = &cobra.Command{
	Use:   "todos",
	Short: "Manage todos",
}

var todosListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List todos in a project",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}

		var kind string
		switch {
		case cmd.Flags().Changed("completed"):
			kind = "completed"
		case cmd.Flags().Changed("pending"):
			kind = "pending"
		default:
			kind = ""
		}

		params := url.Values{}
		params.Add("kind", kind)
		data, err := c.Get(fmt.Sprintf("/projects/%s/todos", projectSid), params)
		if err != nil {
			return err
		}

		printList(data, "todos", nil)
		return nil
	},
}

var todosShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a todo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return notImplemented("GET /todos/:id")
	},
}

var todosCreateCmd = &cobra.Command{
	Use:     "create <title>",
	Short:   "Create a todo",
	Args:    cobra.ExactArgs(1),
	Example: "  $ neetoplanner todos create \"Ship the CLI\" --project engineering",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := resolveProject(cmd); err != nil {
			return err
		}
		return notImplemented("POST /projects/:project_id/todos")
	},
}

var todosUpdateCmd = &cobra.Command{
	Use:   "update <handle>",
	Short: "Update a todo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}
		handle := args[0]

		todo := map[string]interface{}{}
		if cmd.Flags().Changed("title") {
			title, _ := cmd.Flags().GetString("title")
			todo["name"] = title
		}
		if cmd.Flags().Changed("completed") {
			todo["completed"] = true
		}
		if cmd.Flags().Changed("pending") {
			todo["completed"] = false
		}

		if len(todo) == 0 {
			return fmt.Errorf("no fields to update; pass --title, --completed, or --pending")
		}

		body := map[string]interface{}{"todo": todo}
		if _, err = c.Put(fmt.Sprintf("/projects/%s/todos/%s", projectSid, handle), body); err != nil {
			return err
		}

		fmt.Println("Success!")
		return nil
	},
}

var todosDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a todo as completed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return notImplemented("PUT /todos/:id")
	},
}

func init() {
	addProjectFlag(todosListCmd)
	addListFlag(todosListCmd)
	addPaginationFlags(todosListCmd)
	todosListCmd.Flags().Bool("completed", false, "Show the completed todos")
	todosListCmd.Flags().Bool("pending", false, "Show the pending todos")

	addProjectFlag(todosCreateCmd)
	addListFlag(todosCreateCmd)
	todosCreateCmd.Flags().String("assignee", "", "Assignee email")
	todosCreateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")

	addProjectFlag(todosUpdateCmd)
	todosUpdateCmd.Flags().String("title", "", "New title")
	todosUpdateCmd.Flags().String("assignee", "", "Assignee email")
	todosUpdateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
	todosUpdateCmd.Flags().Bool("completed", false, "Mark as completed")
	todosUpdateCmd.Flags().Bool("pending", false, "Mark as pending")
	todosUpdateCmd.MarkFlagsMutuallyExclusive("pending", "completed")

	todosCmd.AddCommand(todosListCmd)
	todosCmd.AddCommand(todosShowCmd)
	todosCmd.AddCommand(todosCreateCmd)
	todosCmd.AddCommand(todosUpdateCmd)
	todosCmd.AddCommand(todosDoneCmd)

	rootCmd.AddCommand(todosCmd)
}
