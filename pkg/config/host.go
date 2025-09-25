package config

import (
	"errors"
	"fmt"
	"os/user"
	"strconv"
	"strings"
)

type InvalidAddressFormat struct {
	address string
}

func (i InvalidAddressFormat) Error() string {
	return fmt.Sprintf("Invalid address format, expected ssh address format '[user@]address[:][port]' but got %s",
		i.address)
}

var ErrInvalidHost = errors.New("Empty host name")
var ErrInvalidPortNum = errors.New("Empty port number")
var ErrInvalidUser = errors.New("Empty username")

type EndpointInfo struct {
	User string
	Host string
	Port uint16
}

func (cfg *Config) ValidateParseHosts() ([]EndpointInfo, error) {
	endpoints := []EndpointInfo{}
	for _, host := range cfg.Hosts {
		ep, err := parseHost(host.Address)
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

func parseHost(host string) (EndpointInfo, error) {
	ep := EndpointInfo{}
	sepUser := strings.Split(host, "@")
	var sepWithoutUser string

	if len(sepUser) == 1 {
		defaultUser, err := user.Current()
		if err != nil {
			return EndpointInfo{}, fmt.Errorf("failed to get current user: %w", err)
		}
		ep.User = defaultUser.Username
		sepWithoutUser = sepUser[0]
	} else {
		ep.User = sepUser[0]
		sepWithoutUser = sepUser[1]
	}
	sepAddr := strings.Split(sepWithoutUser, ":")
	ep.Host = sepAddr[0]
	if len(sepAddr) == 1 {
		ep.Port = 22
	} else {
		port, err := strconv.ParseUint(sepAddr[1], 10, 16)
		if err != nil {
			return EndpointInfo{}, fmt.Errorf("failed to parse port: %w", err)
		}
		ep.Port = uint16(port)
	}
	return ep, nil
}

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
	return nil
}
