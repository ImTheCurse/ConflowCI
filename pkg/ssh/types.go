package ssh

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[SSH]: ", log.Lshortfile|log.LstdFlags)

// Configuration for creating an ssh connection.
type SSHConnConfig struct {
	Username       string
	PrivateKeyPath string
}
