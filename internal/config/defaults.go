package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultsFile = "defaults.json"

type DirFunc func() (string, error)

type Defaults struct {
	DefaultProject string `json:"default_project,omitempty"`
}

type Store struct {
	Subdomains map[string]Defaults `json:"subdomains"`
}

func defaultsFilePath(configDir DirFunc) (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, defaultsFile), nil
}

func Load(configDir DirFunc) (*Store, error) {
	path, err := defaultsFilePath(configDir)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Subdomains: map[string]Defaults{}}, nil
		}
		return nil, fmt.Errorf("Could not read preferences: %w", err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("Invalid preferences file: %w", err)
	}
	if store.Subdomains == nil {
		store.Subdomains = map[string]Defaults{}
	}
	return &store, nil
}

func Save(configDir DirFunc, store *Store) error {
	path, err := defaultsFilePath(configDir)
	if err != nil {
		return err
	}

	if len(store.Subdomains) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("Could not remove preferences: %w", err)
		}
		return nil
	}

	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("Could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("Could not serialize preferences: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("Could not write preferences: %w", err)
	}

	return nil
}

func (s *Store) For(subdomain string) Defaults {
	return s.Subdomains[subdomain]
}

func (s *Store) SetDefaultProject(subdomain, project string) {
	if s.Subdomains == nil {
		s.Subdomains = map[string]Defaults{}
	}
	entry := s.Subdomains[subdomain]
	entry.DefaultProject = project
	s.Subdomains[subdomain] = entry
}

func (s *Store) UnsetDefaultProject(subdomain string) {
	entry, ok := s.Subdomains[subdomain]
	if !ok {
		return
	}
	entry.DefaultProject = ""
	if entry == (Defaults{}) {
		delete(s.Subdomains, subdomain)
		return
	}
	s.Subdomains[subdomain] = entry
}
