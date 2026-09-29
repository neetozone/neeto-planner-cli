package commands

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/client"
	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var (
	app           *cli.App
	registrations []func(root *cobra.Command)
)

func register(fn func(root *cobra.Command)) {
	registrations = append(registrations, fn)
}

func Register(a *cli.App) {
	app = a
	for _, fn := range registrations {
		fn(a.Root())
	}
}

func getClient(cmd *cobra.Command) (*client.Client, error) { return app.Client(cmd) }

func printList(data json.RawMessage, resourceKey string, breadcrumbs []output.Breadcrumb) {
	app.PrintList(data, resourceKey, breadcrumbs)
}

func printResource(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	app.PrintResource(data, breadcrumbs)
}

func printActionResult(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	app.PrintActionResult(data, breadcrumbs)
}

func paginationParams(cmd *cobra.Command) url.Values { return app.PaginationParams(cmd) }

func addPaginationFlags(cmd *cobra.Command) { cli.AddPaginationFlags(0, cmd) }

func printMetadata(data json.RawMessage) {
	type Metadata struct {
		Content json.RawMessage `json:"metadata"`
	}
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return
	}
	if wrapped, err := wrapField("metadata", metadata.Content); err == nil {
		fmt.Println()
		printResource(wrapped, nil)
	}
}

func wrapField(name string, val json.RawMessage) (json.RawMessage, error) {
	if len(val) == 0 {
		return nil, fmt.Errorf("no value present")
	}
	wrapped, err := json.Marshal(map[string]json.RawMessage{name: val})
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

func addVerboseFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("verbose", "v", false, "Display more information")
}

func kindParam(cmd *cobra.Command, names ...string) string {
	for _, name := range names {
		if v, _ := cmd.Flags().GetBool(name); v {
			return name
		}
	}
	return ""
}

func configDir() (string, error) { return app.Auth.ConfigDir() }
