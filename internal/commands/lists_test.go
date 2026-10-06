package commands

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/neetozone/neeto-planner-cli/internal/config"
)

func TestListsCreate_ProjectPrecedenceAndOutputs(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		env     string
		saved   string
		project string
		output  string
	}{
		{"flag and JSON", "flag-project", "env-project", "saved-project", "flag-project", "--json"},
		{"environment and quiet", "", "env-project", "saved-project", "env-project", "--quiet"},
		{"saved default and automatic JSON envelope", "", "", "saved-project", "saved-project", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			root, out := testResourceCommand(t, "lists", "create", func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPost || r.URL.Path != "/api/external/v1/projects/"+tt.project+"/lists" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Session-Token") != "test-session-token" || r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing authentication or JSON header")
				}
				var body map[string]map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(body, map[string]map[string]string{"list": {"name": "Backlog"}}) {
					t.Errorf("unexpected payload: %#v", body)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"sid":"list-sid","name":"Backlog","url":"https://acme.neetoplanner.test/list-sid"}`))
			})
			t.Setenv(projectEnvVar, tt.env)
			store := &config.Store{}
			store.SetDefaultProject("acme", tt.saved)
			if err := config.Save(configDir, store); err != nil {
				t.Fatal(err)
			}
			args := []string{"lists", "create", "Backlog"}
			if tt.flag != "" {
				args = append(args, "--project", tt.flag)
			}
			if tt.output != "" {
				args = append(args, tt.output)
			}
			root.SetArgs(args)
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Errorf("requests = %d, want 1", requests)
			}
			switch tt.output {
			case "--json":
				var result struct {
					Data struct {
						SID  string `json:"sid"`
						Name string `json:"name"`
					} `json:"data"`
				}
				if err := json.Unmarshal(out.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result.Data.SID != "list-sid" || result.Data.Name != "Backlog" {
					t.Errorf("unexpected result: %+v", result)
				}
			case "--quiet":
				if out.String() != "list-sid\n" {
					t.Errorf("quiet output = %q", out.String())
				}
			default:
				for _, value := range []string{"Backlog", "list-sid", "neetoplanner lists list --project saved-project"} {
					if !strings.Contains(out.String(), value) {
						t.Errorf("output missing %q: %q", value, out.String())
					}
				}
			}
		})
	}
}

func TestListsCreate_RequiresANameAndProjectBeforeRequest(t *testing.T) {
	for _, args := range [][]string{{"lists", "create"}, {"lists", "create", "one", "two"}, {"lists", "create", "Backlog"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root, out := testResourceCommand(t, "lists", "create", func(w http.ResponseWriter, r *http.Request) {
				t.Error("invalid arguments should not send a request")
			})
			root.SetArgs(args)
			if err := root.Execute(); err == nil {
				t.Error("expected an error")
			}
			if out.Len() != 0 {
				t.Errorf("unexpected success output: %q", out.String())
			}
		})
	}
}

func TestListsCreate_ReturnsAPIErrorsWithoutSuccessOutput(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			root, out := testResourceCommand(t, "lists", "create", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":"Cannot create this list"}`))
			})
			root.SetArgs([]string{"lists", "create", "", "--project", "project-sid"})
			err := root.Execute()
			if err == nil || !strings.Contains(err.Error(), "Cannot create this list") {
				t.Errorf("expected API error, got %v", err)
			}
			if out.Len() != 0 {
				t.Errorf("unexpected success output: %q", out.String())
			}
		})
	}
}
