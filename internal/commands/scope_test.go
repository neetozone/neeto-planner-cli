package commands

import (
	"strings"
	"testing"
)

func TestPickProject_Precedence(t *testing.T) {
	cases := []struct {
		name                  string
		flag, env, configured string
		want                  string
	}{
		{"flag wins over everything", "from-flag", "from-env", "from-config", "from-flag"},
		{"env wins over config", "", "from-env", "from-config", "from-env"},
		{"config is the last resort", "", "", "from-config", "from-config"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pickProject(tc.flag, tc.env, tc.configured)
			if err != nil {
				t.Fatalf("pickProject(...) returned error: %v", err)
			}
			if got != tc.want {
				t.Errorf("pickProject(%q, %q, %q) = %q, want %q", tc.flag, tc.env, tc.configured, got, tc.want)
			}
		})
	}
}

func TestPickProject_NoneSet(t *testing.T) {
	_, err := pickProject("", "", "")
	if err == nil {
		t.Fatal("pickProject with no values should error")
	}
	if !strings.Contains(err.Error(), "config set default-project") {
		t.Errorf("error should tell the user how to set a default, got: %v", err)
	}
	if !strings.Contains(err.Error(), projectEnvVar) {
		t.Errorf("error should name the env var, got: %v", err)
	}
}

func TestNotImplemented_NamesEndpointAndIssue(t *testing.T) {
	err := notImplemented("GET /projects")
	if !strings.Contains(err.Error(), "GET /projects") {
		t.Errorf("error should name the endpoint, got: %v", err)
	}
	if !strings.Contains(err.Error(), "12675") {
		t.Errorf("error should link the tracking issue, got: %v", err)
	}
}
