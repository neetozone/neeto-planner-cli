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

func testTodosCreate(t *testing.T, handler http.HandlerFunc) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	root := testRoot(t)
	cmd := findCommand(root, "todos", "create")
	resetFlags := func() {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if err := flag.Value.Set(flag.DefValue); err != nil {
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

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Setenv("NEETOPLANNER_BASE_URL", server.URL)
	t.Setenv(projectEnvVar, "")
	var out bytes.Buffer
	app.Printer.Out = &out
	root.SetOut(&out)
	root.SetErr(&out)
	return root, &out
}

func TestTodosCreate_PostsTitleDescriptionListAndIdempotencyKey(t *testing.T) {
	description := "Post on X and LinkedIn.\nCommunity: https://community.neeto.com"
	root, out := testTodosCreate(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/external/v1/projects/marpro-sid/todos" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Session-Token") != "test-session-token" {
			t.Error("missing CLI session token")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("request should contain JSON")
		}
		var body map[string]map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		want := map[string]map[string]string{"todo": {
			"name": "Promote the changelog", "description": description,
			"list_sid": "backlog-sid", "external_idempotency_key": "engage:post-1",
		}}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %#v, want %#v", body, want)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"todo-uuid","handle":42,"name":"Promote the changelog","url":"https://acme.neetoplanner.com/r2/projects/marpro-sid/board/42"}`))
	})

	root.SetArgs([]string{"todos", "create", "Promote the changelog", "--project", "marpro-sid",
		"--list", "backlog-sid", "--description", description, "--idempotency-key", "engage:post-1", "--json"})
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
	root, out := testTodosCreate(t, func(w http.ResponseWriter, r *http.Request) {
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
	root, _ := testTodosCreate(t, func(w http.ResponseWriter, r *http.Request) {
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
			root, out := testTodosCreate(t, func(w http.ResponseWriter, r *http.Request) {
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
	root, _ := testTodosCreate(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("create should not send a request without a project")
	})
	root.SetArgs([]string{"todos", "create", "Ship it"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "No project specified") {
		t.Errorf("expected missing project error, got: %v", err)
	}
}
