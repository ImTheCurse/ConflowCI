package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (cfg *Config) ValidateParseHosts() ([]EndpointInfo, error) {
	endpoints := []EndpointInfo{}
	for _, host := range cfg.Hosts {
		ep, err := parseHost(host.Address)
		ep.Name = host.Name
		ep.PrivateKeyPath = host.PrivateKeyPath
		if err != nil {
			return []EndpointInfo{}, err
		}

		err = ValidateEndpoint(ep)
		if err != nil {
			return []EndpointInfo{}, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, nil
}

// Expands the relative path for each host, returns an error if the file path dosen't exist.
func (cfg *Config) ExpandPrivKeyPath() error {
	for i, host := range cfg.Hosts {
		keyPath := host.PrivateKeyPath
		realPath := os.ExpandEnv(keyPath)

		if _, err := os.Stat(realPath); os.IsNotExist(err) {
			return fmt.Errorf("private key file %s for host %s does not exist", realPath, host.Name)
		}

		cfg.Hosts[i].PrivateKeyPath = realPath
	}
	return nil
}

// Parses a host string into an EndpointInfo struct.
func parseHost(host string) (EndpointInfo, error) {
	ep := EndpointInfo{}
	sepUser := strings.Split(host, ":")

	if len(sepUser) == 1 {
		ep.Host = sepUser[0]
		ep.Port = 0
	} else {
		ep.Host = sepUser[0]
		portStr := sepUser[1]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return EndpointInfo{}, fmt.Errorf("invalid port number: %s", portStr)
		}
		ep.Port = uint16(port)
	}
	return ep, nil
}

// Validates the configuration for the endpoint.
func ValidateEndpoint(ep EndpointInfo) error {
	if ep.User == "" {
		return ErrInvalidUser
	}
	if ep.Host == "" {
		return ErrInvalidHost
	}
	if ep.Port == 0 {
		return ErrInvalidPortNum
	}
	if ep.Name == "" {
		return ErrInvalidHostName
	}
	if ep.PrivateKeyPath == "" {
		return ErrInvalidPrivateKeyPath
	}
	return nil
}

func (ep EndpointInfo) GetEndpointURL() string {
	if ep.Port == 0 {
		return ep.Host
	}
	return fmt.Sprintf("%s:%d", ep.Host, ep.Port)
}
