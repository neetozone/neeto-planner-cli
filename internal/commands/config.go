package commands

import (
	"fmt"

	"github.com/neetozone/neeto-planner-cli/internal/config"
	"github.com/spf13/cobra"
)

const defaultProjectKey = "default-project"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI preferences",
}

var configSetCmd = &cobra.Command{
	Use:     "set <key> <value>",
	Short:   "Set a preference",
	Args:    cobra.ExactArgs(2),
	Example: "  $ neetoplanner config set default-project engineering",
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != defaultProjectKey {
			return unknownConfigKey(args[0])
		}

		subdomain, err := activeSubdomain(cmd)
		if err != nil {
			return err
		}
		store, err := config.Load()
		if err != nil {
			return err
		}

		store.SetDefaultProject(subdomain, args[1])
		if err := config.Save(store); err != nil {
			return err
		}

		fmt.Printf("Set %s to %s for %s.\n", defaultProjectKey, args[1], subdomain)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Show a preference",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != defaultProjectKey {
			return unknownConfigKey(args[0])
		}

		subdomain, err := activeSubdomain(cmd)
		if err != nil {
			return err
		}
		store, err := config.Load()
		if err != nil {
			return err
		}

		value := store.For(subdomain).DefaultProject
		if value == "" {
			return fmt.Errorf("No %s set for %s.", defaultProjectKey, subdomain)
		}

		fmt.Println(value)
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Clear a preference",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != defaultProjectKey {
			return unknownConfigKey(args[0])
		}

		subdomain, err := activeSubdomain(cmd)
		if err != nil {
			return err
		}
		store, err := config.Load()
		if err != nil {
			return err
		}

		store.UnsetDefaultProject(subdomain)
		if err := config.Save(store); err != nil {
			return err
		}

		fmt.Printf("Cleared %s for %s.\n", defaultProjectKey, subdomain)
		return nil
	},
}

func unknownConfigKey(key string) error {
	return fmt.Errorf("Unknown preference %q. Supported keys: %s", key, defaultProjectKey)
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configUnsetCmd)

	rootCmd.AddCommand(configCmd)
}
