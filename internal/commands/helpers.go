package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/neetozone/neeto-planner-cli/internal/auth"
	"github.com/neetozone/neeto-planner-cli/internal/client"
	"github.com/neetozone/neeto-planner-cli/internal/output"
	"github.com/spf13/cobra"
)

func getClient(cmd *cobra.Command) (*client.Client, error) {
	subdomain, _ := cmd.Flags().GetString("subdomain")
	creds, err := auth.SelectCredentials(subdomain)
	if err != nil {
		return nil, err
	}
	return client.New(creds), nil
}

func printList(data json.RawMessage, resourceKey string, breadcrumbs []output.Breadcrumb) {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		output.Print(data, breadcrumbs)
		return
	}

	items, hasItems := parsed[resourceKey]
	pagination := parsed["pagination"]

	if hasItems && pagination != nil {
		output.PrintWithPagination(items, pagination, breadcrumbs)
	} else if hasItems {
		output.Print(items, breadcrumbs)
	} else {
		output.Print(data, breadcrumbs)
	}
}

func printResource(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	output.Print(data, breadcrumbs)
}

func printActionResult(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	output.PrintQuiet(data, breadcrumbs)
}

func paginationParams(cmd *cobra.Command) url.Values {
	params := url.Values{}
	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	client.AddPaginationParams(params, page, pageSize)
	return params
}

func addPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().Int("page", 0, "Page number")
	cmd.Flags().Int("page-size", 0, "Items per page (max 100)")
}

func readJSONFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s: %w", path, err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("Invalid JSON in %s: %w", path, err)
	}
	return result, nil
}

func printMetadata(data json.RawMessage) {
	type Metadata struct {
		Organization json.RawMessage `json:"organization"`
		Project      json.RawMessage `json:"project"`
	}
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return
	}
	if len(metadata.Organization) > 0 && string(metadata.Organization) != "null" {
		fmt.Println()
		printResource(wrapField("organization", metadata.Organization), nil)
	}
	if len(metadata.Project) > 0 && string(metadata.Project) != "null" {
		fmt.Println()
		printResource(wrapField("project", metadata.Project), nil)
	}
	fmt.Println()
}

func wrapField(name string, val json.RawMessage) json.RawMessage {
	if len(val) == 0 {
		return nil
	}
	wrapped, err := json.Marshal(map[string]json.RawMessage{name: val})
	if err != nil {
		return nil
	}
	return wrapped
}

func addVerboseFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("verbose", "v", false, "Display more information")
}
