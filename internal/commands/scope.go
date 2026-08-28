package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/neetozone/neeto-planner-cli/internal/auth"
	"github.com/neetozone/neeto-planner-cli/internal/config"
	"github.com/spf13/cobra"
)

const projectEnvVar = "NEETOPLANNER_PROJECT"

// notImplemented reports a command whose backing endpoint has not shipped yet.
func notImplemented(endpoint string) error {
	return fmt.Errorf(
		"Not implemented yet — waiting on %s.\nTrack progress at https://github.com/neetozone/neeto-planner-web/issues/12675",
		endpoint,
	)
}

// pickProject applies the precedence chain: flag, then environment, then the
// saved default. Values are passed in so the ordering is testable on its own.
func pickProject(flagValue, envValue, configValue string) (string, error) {
	for _, candidate := range []string{flagValue, envValue, configValue} {
		if candidate != "" {
			return candidate, nil
		}
	}
	return "", fmt.Errorf(
		"No project specified. Pass --project, set %s, or run:\n  neetoplanner config set default-project <name-or-sid>",
		projectEnvVar,
	)
}

// activeSubdomain resolves which logged-in subdomain the command applies to.
func activeSubdomain(cmd *cobra.Command) (string, error) {
	override, _ := cmd.Flags().GetString("subdomain")
	creds, err := auth.SelectCredentials(override)
	if err != nil {
		return "", err
	}
	return creds.Subdomain, nil
}

// resolveProject returns the SID a command should act on
func resolveProject(cmd *cobra.Command) (string, error) {
	flagValue, _ := cmd.Flags().GetString("project")
	envValue := os.Getenv(projectEnvVar)

	configValue := ""
	if flagValue == "" && envValue == "" {
		subdomain, err := activeSubdomain(cmd)
		if err != nil {
			return "", err
		}
		store, err := config.Load()
		if err != nil {
			return "", err
		}
		configValue = store.For(subdomain).DefaultProject
	}

	projectValue, err := pickProject(flagValue, envValue, configValue)
	if err != nil {
		return "", err
	}

	return resolveProjectSid(cmd, projectValue)
}

func resolveProjectSid(cmd *cobra.Command, projectValue string) (string, error) {
	c, err := getClient(cmd)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("kind", "all")

	data, err := c.Get("/projects", params)
	if err != nil {
		return "", err
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}

	projectItems, hasItems := parsed["projects"]
	if !hasItems {
		return "", fmt.Errorf("Projects not found (404)")
	}

	var parsedItems []map[string]interface{}
	if err := json.Unmarshal(projectItems, &parsedItems); err != nil {
		return "", err
	}

	for _, elem := range parsedItems {
		sid, _ := elem["sid"].(string)
		if strings.EqualFold(sid, projectValue) {
			return sid, nil
		}
	}

	trimmedValue := strings.TrimSpace(projectValue)
	for _, elem := range parsedItems {
		name, _ := elem["name"].(string)
		if strings.EqualFold(strings.TrimSpace(name), trimmedValue) {
			return elem["sid"].(string), nil
		}
	}
	return "", fmt.Errorf("no project matching %q found", projectValue)
}

func addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String("project", "", "Project name or ID (defaults to "+projectEnvVar+" or the saved default)")
}

func addListFlag(cmd *cobra.Command) {
	cmd.Flags().String("list", "", "List name or ID")
}
