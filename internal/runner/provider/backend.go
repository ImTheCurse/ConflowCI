package provider

import (
	"golang.org/x/crypto/ssh"
)

type RepositoryReader interface {
	Clone(conn *ssh.Client, dir string) (string, error)
	Fetch(conn *ssh.Client, dir string) (string, error)
	CreateWorkTree(conn *ssh.Client, repoDir, wrkTreeRelPath string) error
	RemoveWorkTree(conn *ssh.Client, repoDir, wrkTreeRelPath string) error
}
