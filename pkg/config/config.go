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
