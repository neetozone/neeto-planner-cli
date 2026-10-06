package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
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

		kind := kindParam(cmd, "completed", "pending")
		params := paginationParams(cmd)
		params.Add("kind", kind)
		data, err := c.Get(fmt.Sprintf("/projects/%s/todos", projectSid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects ls"},
			{Label: "List lists", Command: "neetoplanner lists ls"},
			{Label: "Show list", Command: "neetoplanner lists show <sid>"},
		}

		printList(data, "todos", breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
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
	Example: "  $ neetoplanner todos create \"Promote the changelog\" --project <project-sid> --list <list-sid>",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}

		todo := map[string]interface{}{"name": args[0]}
		for flag, field := range map[string]string{
			"description":     "description",
			"list":            "list_sid",
			"idempotency-key": "external_idempotency_key",
			"assignee":        "assignee_email",
			"due":             "due_date",
		} {
			if cmd.Flags().Changed(flag) {
				value, _ := cmd.Flags().GetString(flag)
				todo[field] = value
			}
		}

		data, err := c.Post(fmt.Sprintf("/projects/%s/todos", projectSid), map[string]interface{}{"todo": todo})
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List todos", Command: fmt.Sprintf("neetoplanner todos list --project %s", projectSid)},
		}
		printActionResult(data, breadcrumbs)
		return nil
	},
}

var todosUpdateCmd = &cobra.Command{
	Use:   "update <id>",
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
		id := args[0]

		todo := map[string]interface{}{}
		for flag, field := range map[string]string{
			"title":    "name",
			"assignee": "assignee_email",
			"due":      "due_date",
		} {
			if !cmd.Flags().Changed(flag) {
				continue
			}
			value, _ := cmd.Flags().GetString(flag)
			todo[field] = value
		}
		if cmd.Flags().Changed("completed") {
			todo["completed"] = true
		}
		if cmd.Flags().Changed("pending") {
			todo["completed"] = false
		}

		if len(todo) == 0 {
			return fmt.Errorf("No fields to update. Pass --title, --assignee, --due, --completed, or --pending.")
		}

		body := map[string]interface{}{"todo": todo}
		data, err := c.Put(fmt.Sprintf("/projects/%s/todos/%s", projectSid, id), body)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects ls"},
			{Label: "List lists", Command: "neetoplanner lists ls"},
			{Label: "Show list", Command: "neetoplanner lists show <sid>"},
		}

		printResource(data, breadcrumbs)
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
	addPaginationFlags(todosListCmd)
	addVerboseFlag(todosListCmd)
	todosListCmd.Flags().Bool("completed", false, "Show the completed todos")
	todosListCmd.Flags().Bool("pending", false, "Show the pending todos")

	addProjectFlag(todosCreateCmd)
	addListFlag(todosCreateCmd)
	todosCreateCmd.Flags().String("description", "", "Todo description")
	todosCreateCmd.Flags().String("idempotency-key", "", "Stable event key to reuse when retrying creation")
	todosCreateCmd.Flags().String("assignee", "", "Email of the project member to assign")
	todosCreateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD; empty skips the dependency mode default)")

	addProjectFlag(todosUpdateCmd)
	todosUpdateCmd.Flags().String("title", "", "New title")
	todosUpdateCmd.Flags().String("assignee", "", "Replace assignees with this project member's email (empty clears)")
	todosUpdateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD; empty clears)")
	todosUpdateCmd.Flags().Bool("completed", false, "Mark as completed")
	todosUpdateCmd.Flags().Bool("pending", false, "Mark as pending")
	todosUpdateCmd.MarkFlagsMutuallyExclusive("pending", "completed")

	todosCmd.AddCommand(todosListCmd)
	todosCmd.AddCommand(todosShowCmd)
	todosCmd.AddCommand(todosCreateCmd)
	todosCmd.AddCommand(todosUpdateCmd)
	todosCmd.AddCommand(todosDoneCmd)

	register(func(root *cobra.Command) { root.AddCommand(todosCmd) })
}
