package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	if err := os.WriteFile(path, []byte(`{"name":"test","count":42}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := readJSONFile(path)
	if err != nil {
		t.Fatalf("readJSONFile() error = %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["count"] != float64(42) {
		t.Errorf("count = %v, want 42", result["count"])
	}
}

func TestReadJSONFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{not valid`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := readJSONFile(path)
	if err == nil {
		t.Error("readJSONFile() expected error for invalid JSON")
	}
}

func TestReadJSONFile_NotFound(t *testing.T) {
	_, err := readJSONFile("/nonexistent/file.json")
	if err == nil {
		t.Error("readJSONFile() expected error for missing file")
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() error = %v", err)
	}
	os.Stdout = orig

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	return buf.String()
}

func TestPrintOrganizationInformation_Valid(t *testing.T) {
	data := json.RawMessage(`{"organization":"Acme Inc"}`)
	out := captureStdout(t, func() { printOrganizationInformation(data) })
	want := "Organization: Acme Inc\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintOrganizationInformation_Missing(t *testing.T) {
	data := json.RawMessage(`{"foo":"bar"}`)
	out := captureStdout(t, func() { printOrganizationInformation(data) })
	want := "Organization: -\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintOrganizationInformation_WrongType(t *testing.T) {
	data := json.RawMessage(`{"organization":123}`)
	out := captureStdout(t, func() { printOrganizationInformation(data) })
	want := "Organization: -\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintOrganizationInformation_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)
	out := captureStdout(t, func() { printOrganizationInformation(data) })
	if out != "" {
		t.Errorf("output = %q, want empty on invalid JSON", out)
	}
}

func TestPrintProjectInformation_Valid(t *testing.T) {
	data := json.RawMessage(`{"project":{"name":"Engineering","sid":"proj_123"}}`)
	out := captureStdout(t, func() { printProjectInformation(data) })
	want := "Project: Engineering (proj_123)\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintProjectInformation_MissingProject(t *testing.T) {
	data := json.RawMessage(`{"organization":"Acme"}`)
	out := captureStdout(t, func() { printProjectInformation(data) })
	want := "Project: -\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintProjectInformation_ProjectWrongType(t *testing.T) {
	data := json.RawMessage(`{"project":"not-an-object"}`)
	out := captureStdout(t, func() { printProjectInformation(data) })
	want := "Project: -\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintProjectInformation_NameMissing(t *testing.T) {
	data := json.RawMessage(`{"project":{"sid":"proj_123"}}`)
	out := captureStdout(t, func() { printProjectInformation(data) })
	want := "Project: - (proj_123)\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintProjectInformation_SidMissing(t *testing.T) {
	data := json.RawMessage(`{"project":{"name":"Engineering"}}`)
	out := captureStdout(t, func() { printProjectInformation(data) })
	want := "Project: Engineering (-)\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintProjectInformation_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)
	out := captureStdout(t, func() { printProjectInformation(data) })
	if out != "" {
		t.Errorf("output = %q, want empty on invalid JSON", out)
	}
}

func TestPrintTotalCount_Valid(t *testing.T) {
	data := json.RawMessage(`{"total_count":17}`)
	out := captureStdout(t, func() { printTotalCount(data) })
	want := "Total count: 17\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintTotalCount_Zero(t *testing.T) {
	data := json.RawMessage(`{"total_count":0}`)
	out := captureStdout(t, func() { printTotalCount(data) })
	want := "Total count: 0\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintTotalCount_Missing(t *testing.T) {
	data := json.RawMessage(`{"foo":"bar"}`)
	out := captureStdout(t, func() { printTotalCount(data) })
	want := "Total count: -\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintTotalCount_WrongType(t *testing.T) {
	data := json.RawMessage(`{"total_count":"seventeen"}`)
	out := captureStdout(t, func() { printTotalCount(data) })
	want := "Total count: -\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPrintTotalCount_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)
	out := captureStdout(t, func() { printTotalCount(data) })
	if out != "" {
		t.Errorf("output = %q, want empty on invalid JSON", out)
	}
}

func TestNewLine(t *testing.T) {
	out := captureStdout(t, func() { newLine() })
	if out != "\n" {
		t.Errorf("output = %q, want %q", out, "\n")
	}
}
