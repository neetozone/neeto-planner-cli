package commands

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-planner-cli"
)

func testRoot(t *testing.T) *cobra.Command {
	t.Helper()
	cfg, err := config.Parse(product.ConfigYAML)
	if err != nil {
		t.Fatalf("config.Parse: %v", err)
	}
	a := cli.New(*cfg)
	Register(a)
	return a.Root()
}

func findCommand(root *cobra.Command, path ...string) *cobra.Command {
	current := root
	for _, name := range path {
		found := false
		for _, child := range current.Commands() {
			if child.Name() == name {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

func TestCommandTree_CoversEpicEndpoints(t *testing.T) {
	paths := [][]string{
		{"projects", "list"},
		{"lists", "list"},
		{"lists", "create"},
		{"todos", "list"},
		{"todos", "show"},
		{"todos", "create"},
		{"todos", "update"},
		{"todos", "done"},
		{"config", "set"},
	}

	root := testRoot(t)
	for _, path := range paths {
		if findCommand(root, path...) == nil {
			t.Errorf("command %v is not registered", path)
		}
	}
}

func TestScopedCommands_HaveProjectFlag(t *testing.T) {
	paths := [][]string{
		{"lists", "list"},
		{"lists", "create"},
		{"todos", "list"},
		{"todos", "create"},
	}

	root := testRoot(t)
	for _, path := range paths {
		cmd := findCommand(root, path...)
		if cmd == nil {
			t.Fatalf("command %v is not registered", path)
		}
		if cmd.Flags().Lookup("project") == nil {
			t.Errorf("command %v should accept --project", path)
		}
	}
}

func TestTodosCreate_TakesPositionalTitle(t *testing.T) {
	cmd := findCommand(testRoot(t), "todos", "create")
	if cmd == nil {
		t.Fatal("todos create is not registered")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("todos create should reject a missing title")
	}
	if err := cmd.Args(cmd, []string{"Ship the CLI"}); err != nil {
		t.Errorf("todos create should accept a single positional title, got: %v", err)
	}
}
