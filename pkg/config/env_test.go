package config

import (
	"os"
	"testing"
)

func TestExpandEnv(t *testing.T) {
	// Set up environment variable for token substitution
	err := os.Setenv("GITHUB_TOKEN", "test-token-123")
	if err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("GITHUB_TOKEN")

	cfg := &Config{
		Env: &Environment{
			GlobalEnv: map[string]string{
				"GITHUB_TOKEN": "${GITHUB_TOKEN}",
			},
		},
	}

	err = cfg.expandEnv()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.Env.GlobalEnv["GITHUB_TOKEN"] != "test-token-123" {
		t.Errorf("expected GITHUB_TOKEN to be test-token-123, got %s", cfg.Env.GlobalEnv["GITHUB_TOKEN"])
	}

}
