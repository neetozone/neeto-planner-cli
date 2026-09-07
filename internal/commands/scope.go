package commands

import (
	"fmt"
	"os"

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
	creds, err := app.Auth.SelectCredentials(override)
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
		store, err := config.Load(configDir)
		if err != nil {
			return "", err
		}
		configValue = store.For(subdomain).DefaultProject
	}

	return pickProject(flagValue, envValue, configValue)
}

func addProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String("project", "", "Project name or ID (defaults to "+projectEnvVar+" or the saved default)")
}

func addListFlag(cmd *cobra.Command) {
	cmd.Flags().String("list", "", "List name or ID")
}
