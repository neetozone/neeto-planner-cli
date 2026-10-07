package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/neetozone/neeto-cli-commons/auth"
	"github.com/neetozone/neeto-planner-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type todosTransport func(*http.Request) (*http.Response, error)

func (transport todosTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func testTodosCommand(t *testing.T, command string, handler http.HandlerFunc) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	return testResourceCommand(t, "todos", command, handler)
}

func testResourceCommand(t *testing.T, resource, command string, handler http.HandlerFunc) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	root := testRoot(t)
	cmd := findCommand(root, resource, command)
	resetFlags := func() {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			var err error
			if slice, ok := flag.Value.(pflag.SliceValue); ok {
				err = slice.Replace([]string{})
			} else {
				err = flag.Value.Set(flag.DefValue)
			}
			if err != nil {
				t.Fatalf("reset %s: %v", flag.Name, err)
			}
			flag.Changed = false
		})
	}
	resetFlags()
	t.Cleanup(resetFlags)

	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg := app.Product
	cfg.ConfigDir, err = filepath.Rel(userHome, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app.Auth = auth.New(cfg)
	if err := app.Auth.SaveStore(&auth.Store{Credentials: []auth.Credentials{{
		Subdomain: "acme", Email: "marketing@example.com", SessionToken: "test-session-token",
	}}}); err != nil {
		t.Fatal(err)
	}

	previousTransport := http.DefaultTransport
	http.DefaultTransport = todosTransport(func(request *http.Request) (*http.Response, error) {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Result(), nil
	})
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	t.Setenv("NEETOPLANNER_BASE_URL", "https://acme.neetoplanner.test")
	t.Setenv(projectEnvVar, "")
	var out bytes.Buffer
	app.Printer.Out = &out
	root.SetOut(&out)
	root.SetErr(&out)
	return root, &out
}

func TestTodosCreate_PostsRequestedFields(t *testing.T) {
	description := "Post on X and LinkedIn.\nCommunity: https://community.neeto.com"
	root, out := testTodosCommand(t, "create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/external/v1/projects/marpro-sid/todos" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Session-Token") != "test-session-token" {
			t.Error("missing CLI session token")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("request should contain JSON")
		}
		var body map[string]map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		want := map[string]map[string]any{"todo": {
			"name": "Promote the changelog", "description": description,
			"list_sid": "backlog-sid", "external_idempotency_key": "engage:post-1",
			"assignee_email": "member@example.com", "due_date": "2028-02-29",
			"tags": []any{"Urgent", "Review"},
		}}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %#v, want %#v", body, want)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"todo-uuid","handle":42,"name":"Promote the changelog","url":"https://acme.neetoplanner.com/r2/projects/marpro-sid/board/42"}`))
	})

	root.SetArgs([]string{"todos", "create", "Promote the changelog", "--project", "marpro-sid",
		"--list", "backlog-sid", "--description", description, "--idempotency-key", "engage:post-1",
		"--assignee", "member@example.com", "--due", "2028-02-29", "--tags", "Urgent,Review", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Data struct {
			ID     string `json:"id"`
			Handle int    `json:"handle"`
			URL    string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if result.Data.ID != "todo-uuid" || result.Data.Handle != 42 || result.Data.URL == "" {
		t.Errorf("unexpected create result: %+v", result.Data)
	}
}

func TestTodosCreate_UsesEnvironmentProjectAndPrintsQuietIdentifier(t *testing.T) {
	root, out := testTodosCommand(t, "create", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/external/v1/projects/from-env/todos" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if !reflect.DeepEqual(body, map[string]map[string]string{"todo": {"name": "Ship it"}}) {
			t.Errorf("unset options should be omitted: %#v", body)
		}
		_, _ = w.Write([]byte(`{"id":"todo-uuid","name":"Ship it"}`))
	})
	t.Setenv(projectEnvVar, "from-env")
	root.SetArgs([]string{"todos", "create", "Ship it", "--quiet"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "todo-uuid\n" {
		t.Errorf("quiet output = %q, want only the todo identifier", out.String())
	}
}

func TestTodosCreate_UsesSavedDefaultProject(t *testing.T) {
	root, _ := testTodosCommand(t, "create", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/external/v1/projects/saved-project/todos" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"todo-uuid"}`))
	})
	store := &config.Store{}
	store.SetDefaultProject("acme", "saved-project")
	if err := config.Save(configDir, store); err != nil {
		t.Fatal(err)
	}
	root.SetArgs([]string{"todos", "create", "Ship it"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestTodosCreate_ReturnsAPIErrorsWithoutSuccessOutput(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusUnprocessableEntity} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			root, out := testTodosCommand(t, "create", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":"Cannot create this todo"}`))
			})
			root.SetArgs([]string{"todos", "create", "Ship it", "--project", "marpro-sid"})
			err := root.Execute()
			if err == nil || !strings.Contains(err.Error(), "Cannot create this todo") {
				t.Errorf("create should return the API error, got: %v", err)
			}
			if out.Len() != 0 {
				t.Errorf("failed create should not print success: %q", out.String())
			}
		})
	}
}

func TestTodosCreate_RequiresAProjectBeforeSendingARequest(t *testing.T) {
	root, _ := testTodosCommand(t, "create", func(w http.ResponseWriter, r *http.Request) {
		t.Error("create should not send a request without a project")
	})
	root.SetArgs([]string{"todos", "create", "Ship it"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "No project specified") {
		t.Errorf("expected missing project error, got: %v", err)
	}
}

func TestTodosUpdate_SendsOnlyRequestedChanges(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  map[string]any
	}{
		{
			name:  "assignee",
			flags: []string{"--assignee", "member@example.com"},
			want:  map[string]any{"assignee_email": "member@example.com"},
		},
		{
			name:  "due date",
			flags: []string{"--due", "2028-02-29"},
			want:  map[string]any{"due_date": "2028-02-29"},
		},
		{
			name:  "clear fields",
			flags: []string{"--assignee", "", "--due", ""},
			want:  map[string]any{"assignee_email": "", "due_date": ""},
		},
		{
			name:  "multiple tags",
			flags: []string{"--tags", "Urgent,Review", "--tags", "Urgent"},
			want:  map[string]any{"tags": []any{"Urgent", "Review", "Urgent"}},
		},
		{
			name:  "clear tags",
			flags: []string{"--tags", ""},
			want:  map[string]any{"tags": []any{}},
		},
		{
			name: "combined changes",
			flags: []string{
				"--title", "Updated", "--assignee", "member@example.com", "--due", "2026-10-10", "--completed",
				"--tags", "Urgent",
			},
			want: map[string]any{
				"name": "Updated", "assignee_email": "member@example.com",
				"due_date": "2026-10-10", "completed": true,
				"tags": []any{"Urgent"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			root, out := testTodosCommand(t, "update", func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPut || r.URL.Path != "/api/external/v1/projects/project-sid/todos/42" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var body map[string]map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if !reflect.DeepEqual(body, map[string]map[string]any{"todo": tt.want}) {
					t.Errorf("body = %#v, want only %#v", body, tt.want)
				}
				_, _ = w.Write([]byte(`{"handle":42,"name":"Updated"}`))
			})
			args := []string{"todos", "update", "42", "--project", "project-sid", "--json"}
			root.SetArgs(append(args, tt.flags...))
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Errorf("requests = %d, want 1", requests)
			}
			if !json.Valid(out.Bytes()) {
				t.Errorf("output is not JSON: %q", out.String())
			}
		})
	}
}

func TestTodosCreate_SendsEmptyTags(t *testing.T) {
	root, _ := testTodosCommand(t, "create", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		want := map[string]map[string]any{"todo": {"name": "Ship it", "tags": []any{}}}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %#v, want %#v", body, want)
		}
		_, _ = w.Write([]byte(`{"id":"todo-uuid"}`))
	})
	root.SetArgs([]string{"todos", "create", "Ship it", "--project", "project-sid", "--tags", ""})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}
