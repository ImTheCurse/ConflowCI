package ssh

import (
	"fmt"
	"os"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

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
		logger.Println("Built SSH config.")
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
