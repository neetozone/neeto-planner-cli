package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
}

var tasksListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List tasks in a project",
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
		data, err := c.Get(fmt.Sprintf("/projects/%s/tasks", projectSid), params)
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List projects", Command: "neetoplanner projects ls"},
			{Label: "List lists", Command: "neetoplanner lists ls"},
			{Label: "Show list", Command: "neetoplanner lists show <sid>"},
		}

		printList(data, "tasks", breadcrumbs)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printMetadata(data)
		}
		return nil
	},
}

var tasksShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return notImplemented("GET /tasks/:id")
	},
}

var tasksCreateCmd = &cobra.Command{
	Use:     "create <title>",
	Short:   "Create a task",
	Args:    cobra.ExactArgs(1),
	Example: "  $ neetoplanner tasks create \"Promote the changelog\" --project <project-sid> --list <list-sid>",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}
		projectSid, err := resolveProject(cmd)
		if err != nil {
			return err
		}

		task := map[string]interface{}{"name": args[0]}
		for flag, field := range map[string]string{
			"description":     "description",
			"list":            "list_sid",
			"idempotency-key": "external_idempotency_key",
			"assignee":        "assignee_email",
			"due":             "due_date",
		} {
			if cmd.Flags().Changed(flag) {
				value, _ := cmd.Flags().GetString(flag)
				task[field] = value
			}
		}

		if cmd.Flags().Changed("tags") {
			task["tags"], _ = cmd.Flags().GetStringSlice("tags")
		}

		data, err := c.Post(fmt.Sprintf("/projects/%s/tasks", projectSid), map[string]interface{}{"task": task})
		if err != nil {
			return err
		}

		breadcrumbs := []output.Breadcrumb{
			{Label: "List tasks", Command: fmt.Sprintf("neetoplanner tasks list --project %s", projectSid)},
		}
		printActionResult(data, breadcrumbs)
		return nil
	},
}

var tasksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a task",
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

		task := map[string]interface{}{}
		for flag, field := range map[string]string{
			"title":    "name",
			"assignee": "assignee_email",
			"due":      "due_date",
		} {
			if !cmd.Flags().Changed(flag) {
				continue
			}
			value, _ := cmd.Flags().GetString(flag)
			task[field] = value
		}
		if cmd.Flags().Changed("completed") {
			task["completed"] = true
		}
		if cmd.Flags().Changed("pending") {
			task["completed"] = false
		}
		if cmd.Flags().Changed("tags") {
			task["tags"], _ = cmd.Flags().GetStringSlice("tags")
		}

		if len(task) == 0 {
			return fmt.Errorf("No fields to update. Pass --title, --assignee, --due, --tags, --completed, or --pending.")
		}

		body := map[string]interface{}{"task": task}
		data, err := c.Put(fmt.Sprintf("/projects/%s/tasks/%s", projectSid, id), body)
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

var tasksDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return notImplemented("PUT /tasks/:id")
	},
}

func init() {
	addProjectFlag(tasksListCmd)
	addPaginationFlags(tasksListCmd)
	addVerboseFlag(tasksListCmd)
	tasksListCmd.Flags().Bool("completed", false, "Show the completed tasks")
	tasksListCmd.Flags().Bool("pending", false, "Show the pending tasks")

	addProjectFlag(tasksCreateCmd)
	addListFlag(tasksCreateCmd)
	tasksCreateCmd.Flags().String("description", "", "Task description")
	tasksCreateCmd.Flags().String("idempotency-key", "", "Stable event key to reuse when retrying creation")
	tasksCreateCmd.Flags().String("assignee", "", "Email of the project member to assign")
	tasksCreateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD; empty skips the dependency mode default)")
	tasksCreateCmd.Flags().StringSlice("tags", []string{}, "Existing project tag names, separated by commas")

	addProjectFlag(tasksUpdateCmd)
	tasksUpdateCmd.Flags().String("title", "", "New title")
	tasksUpdateCmd.Flags().String("assignee", "", "Replace assignees with this project member's email (empty clears)")
	tasksUpdateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD; empty clears)")
	tasksUpdateCmd.Flags().StringSlice("tags", []string{}, "Project tag names (comma-separated; replaces tags; empty clears)")
	tasksUpdateCmd.Flags().Bool("completed", false, "Mark as completed")
	tasksUpdateCmd.Flags().Bool("pending", false, "Mark as pending")
	tasksUpdateCmd.MarkFlagsMutuallyExclusive("pending", "completed")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksShowCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCmd.AddCommand(tasksUpdateCmd)
	tasksCmd.AddCommand(tasksDoneCmd)

	register(func(root *cobra.Command) { root.AddCommand(tasksCmd) })
}
