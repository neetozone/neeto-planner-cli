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

func printOrganizationInformation(data json.RawMessage) {
	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		fmt.Println("error decoding data")
		return
	}
	organization, ok := metadata["organization"].(string)
	if !ok {
		organization = "-"
	}
	fmt.Printf("Organization: %s\n", organization)
}

func printProjectInformation(data json.RawMessage) {
	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		fmt.Println("error decoding data")
		return
	}
	project, ok := metadata["project"].(map[string]interface{})
	if !ok {
		fmt.Printf("Project: -\n")
		return
	}
	name, ok := project["name"].(string)
	if !ok {
		name = "-"
	}
	sid, ok := project["sid"].(string)
	if !ok {
		sid = "-"
	}
	fmt.Printf("Project: %s (%s)\n", name, sid)
}

func printTotalCount(data json.RawMessage) {
	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		fmt.Println("error decoding data")
		return
	}

	switch v := metadata["total_count"].(type) {
	case float64:
		fmt.Printf("Total count: %d\n", int(v))
	default:
		fmt.Printf("Total count: -\n")
	}
}

func newLine() {
	fmt.Println()
}
