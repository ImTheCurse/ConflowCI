package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var logger = log.New(os.Stdout, "[Config Parser]: ", log.Lshortfile|log.LstdFlags)

func NewConfig(filename string) *Config {
	cfg := &Config{}

	b, err := os.ReadFile(filename)
	if err != nil {
		panic("Couldn't read config file, Make sure it exist and has read permissions.")
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		panic("Couldn't parse config file, make sure the config file has valid yaml format and required fields exist.")
	}
	return cfg
}
