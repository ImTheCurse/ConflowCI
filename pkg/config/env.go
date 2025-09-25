package config

import (
	"fmt"
	"os"
)

// ExpandEnv expands environment variables in the config.
func (cfg *Config) expandEnv() error {
	if cfg == nil {
		return fmt.Errorf("Config is nil.")
	}
	if cfg.Provider.Github.Auth != nil {
		token := cfg.Provider.Github.Auth.Token
		expandedVar := os.ExpandEnv(token)
		if len(expandedVar) <= 0 {
			return fmt.Errorf("Environment variable for Github auth token dosen't exist")
		}
		cfg.Provider.Github.Auth.Token = expandedVar
	}

	if cfg.Env == nil {
		return nil
	}

	if cfg.Env.GlobalEnv != nil {
		for key, val := range cfg.Env.GlobalEnv {
			expandedVar := os.ExpandEnv(val)
			if len(expandedVar) <= 0 {
				return fmt.Errorf("global environment variable for key %v dosen't exist", key)
			}
			cfg.Env.GlobalEnv[key] = expandedVar
		}
	}

	if cfg.Env.LocalEnv == nil {
		return nil
	}

	for key, val := range cfg.Env.LocalEnv {
		expandedVar := os.ExpandEnv(val)
		if len(expandedVar) <= 0 {
			return fmt.Errorf("local environment variable for key %v dosen't exist", key)
		}
		cfg.Env.LocalEnv[key] = expandedVar
	}
	return nil
}
