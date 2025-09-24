package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var logger = log.New(os.Stdout, "[Config Parser]: ", log.Lshortfile|log.LstdFlags)

// NewConfig creates a new Config instance from a YAML file, the function also expands environment variables for the Enviornmet
// field and the auth field in the github provider.
func NewConfig(filename string) (*Config, error) {
	cfg := &Config{}

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read config file, Make sure it exist and has read permissions.")
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse config file, make sure the config file has valid yaml format and required fields exist.")
	}

	err = cfg.expandEnv()
	if err != nil {
		return nil, err
	}
	return cfg, err
}

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
