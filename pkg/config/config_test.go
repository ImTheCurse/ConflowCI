package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Set up environment variable for token substitution
	err := os.Setenv("GITHUB_TOKEN", "test-token-123")
	if err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("GITHUB_TOKEN")
	err = os.Setenv("CONFLOW_TEST_DIR", "testdata")
	if err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("CONFLOW_TEST_DIR")

	// Load the test YAML file
	yamlPath := filepath.Join("testdata", "test-config.yaml")

	config, err := NewConfig(yamlPath)
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	if config.Provider.Github.Repository != "org/repo-name" {
		t.Errorf("Expected repository 'org/repo-name', got '%s'", config.Provider.Github.Repository)
	}

	if config.Provider.Github.Auth.Token != "test-token-123" {
		t.Errorf("Expected token 'test-token-123', got '%s'", config.Provider.Github.Auth.Token)
	}

	if len(config.Hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(config.Hosts))
	}
	priv_key_path := "./testdata/fake_id_rsa"
	for _, host := range config.Hosts {
		if host.PrivateKeyPath != priv_key_path {
			t.Errorf("Expected private key path '%s', got '%s'", priv_key_path, host.PrivateKeyPath)
		}
	}
}
