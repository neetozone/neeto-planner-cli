package config

import (
	"os"
	"testing"
)

func TestStore_SetAndUnsetDefaultProject(t *testing.T) {
	store := &Store{Subdomains: map[string]Defaults{}}

	store.SetDefaultProject("acme", "engineering")
	if got := store.For("acme").DefaultProject; got != "engineering" {
		t.Errorf("For(\"acme\").DefaultProject = %q, want engineering", got)
	}

	store.SetDefaultProject("other", "marketing")
	if got := store.For("acme").DefaultProject; got != "engineering" {
		t.Errorf("setting another subdomain changed acme: got %q", got)
	}

	store.UnsetDefaultProject("acme")
	if got := store.For("acme").DefaultProject; got != "" {
		t.Errorf("after unset, DefaultProject = %q, want empty", got)
	}
	if got := store.For("other").DefaultProject; got != "marketing" {
		t.Errorf("unset leaked across subdomains: other = %q", got)
	}
}

func TestStore_ForUnknownSubdomain(t *testing.T) {
	store := &Store{Subdomains: map[string]Defaults{}}
	if got := store.For("nobody").DefaultProject; got != "" {
		t.Errorf("unknown subdomain should yield zero value, got %q", got)
	}
}

func TestLoadAndSave_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := Load()
	if err != nil {
		t.Fatalf("Load on a fresh home: %v", err)
	}
	if len(store.Subdomains) != 0 {
		t.Errorf("fresh store should be empty, got %d entries", len(store.Subdomains))
	}

	store.SetDefaultProject("acme", "engineering")
	if err := Save(store); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if got := reloaded.For("acme").DefaultProject; got != "engineering" {
		t.Errorf("round-tripped DefaultProject = %q, want engineering", got)
	}
}

func TestSave_RemovesFileWhenEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store := &Store{Subdomains: map[string]Defaults{}}
	store.SetDefaultProject("acme", "engineering")
	if err := Save(store); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path, err := defaultsFilePath()
	if err != nil {
		t.Fatalf("defaultsFilePath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("preferences file should exist after Save: %v", err)
	}

	store.UnsetDefaultProject("acme")
	if err := Save(store); err != nil {
		t.Fatalf("Save after unset: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("preferences file should be removed when empty, stat err = %v", err)
	}
}

func TestSave_FileIsNotWorldReadable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store := &Store{Subdomains: map[string]Defaults{}}
	store.SetDefaultProject("acme", "engineering")
	if err := Save(store); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path, err := defaultsFilePath()
	if err != nil {
		t.Fatalf("defaultsFilePath: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("preferences file mode = %o, want 600", perm)
	}
}
