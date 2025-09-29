package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// NewConfig creates a new validated Config instance from a YAML file,
// the function also expands environment variables for the Enviornmet
// field and the auth field in the github provider.
func NewConfig(filename string) (*ValidatedConfig, error) {
	cfg := &Config{}

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read config file, Make sure it exist and has read permissions.")
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse config file, make sure the config file has valid yaml format and required fields exist.")
	}
	logger.Println("Config file parsed successfully, expanding env...")
	err = cfg.expandEnv()
	if err != nil {
		return nil, err
	}
	err = cfg.ExpandPrivKeyPath()
	if err != nil {
		return nil, err
	}
	logger.Println("Expanded env, validating config fields...")
	cfg.ValidatePipeline()
	cfg.ValidateProvider()
	eps, err := cfg.ValidateParseHosts()
	if err != nil {
		return nil, err
	}
	logger.Println("Finished config validation.")
	validatedCfg := &ValidatedConfig{
		Config:    cfg,
		Endpoints: eps,
	}
	return validatedCfg, err
}
