package ssh

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var logger = log.New(os.Stdout, "[SSH]: ", log.Lshortfile|log.LstdFlags)

var ErrNotSupported = errors.New("Authentication method not supported")
var ErrPrivKetFileNotFound = errors.New("Public key file was not found")
var ErrPrivateKeyParse = errors.New("Private key file could not be parsed")
var ErrEmptyPrivKeyPath = errors.New("Private key path is empty")

// Configuration for creating an ssh connection.
type SSHConnConfig struct {
	Username       string
	PrivateKeyPath string
}

// BuildConfig builds a new SSH client configuration using username, and a path to
// the private key that the SSH server authenticates against.
func (s SSHConnConfig) BuildConfig() (*ssh.ClientConfig, error) {
	if len(s.PrivateKeyPath) > 0 {
		key, err := os.ReadFile(s.PrivateKeyPath)
		if err != nil {
			return nil, ErrPrivKetFileNotFound
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, ErrPrivateKeyParse
		}
		// define auth method
		auth := ssh.PublicKeys(signer)

		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		hostKeyCallback, err := knownhosts.New(fmt.Sprintf("%s/.ssh/known_hosts", userHomeDir))
		if err != nil {
			return nil, err
		}
		return &ssh.ClientConfig{
			User: s.Username,
			Auth: []ssh.AuthMethod{
				auth,
			},
			HostKeyCallback: hostKeyCallback,
		}, nil
	} else {
		return nil, ErrEmptyPrivKeyPath
	}
}

// Creates a new SSH connection using the provided configuration.
// you need to have an SSH config. you can use the ssh.BuildConfig function
// in order create it.
func NewSSHConn(ep config.EndpointInfo, cfg *ssh.ClientConfig) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", ep.Host, ep.Port)
	logger.Println("Starting SSH connection")
	conn, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
