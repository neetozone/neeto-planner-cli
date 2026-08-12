package commands

import (
	"github.com/spf13/cobra"
)

var todosCmd = &cobra.Command{
	Use:   "todos",
	Short: "Manage todos",
}

var todosListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List todos in a project or list",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := resolveProject(cmd); err != nil {
			return err
		}
		if list, _ := cmd.Flags().GetString("list"); list != "" {
			return notImplemented("GET /projects/:project_id/lists/:list_id/todos")
		}
		return notImplemented("GET /projects/:project_id/todos")
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
	Use:   "update <id>",
	Short: "Update a todo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return notImplemented("PUT /todos/:id")
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

	addProjectFlag(todosCreateCmd)
	addListFlag(todosCreateCmd)
	todosCreateCmd.Flags().String("assignee", "", "Assignee email")
	todosCreateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")

	todosUpdateCmd.Flags().String("title", "", "New title")
	todosUpdateCmd.Flags().String("assignee", "", "Assignee email")
	todosUpdateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
	todosUpdateCmd.Flags().Bool("completed", false, "Mark as completed")

	todosCmd.AddCommand(todosListCmd)
	todosCmd.AddCommand(todosShowCmd)
	todosCmd.AddCommand(todosCreateCmd)
	todosCmd.AddCommand(todosUpdateCmd)
	todosCmd.AddCommand(todosDoneCmd)

	rootCmd.AddCommand(todosCmd)
}
