package ssh

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
)

func TestNewSSHConn(t *testing.T) {
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Errorf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")
	ctx := context.Background()
	container, err := CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Errorf("Failed to start SSH server container: %v", err)
	}
	fmt.Println("SSH server running at", Ep.Host, Ep.Port)
	defer container.Terminate(ctx)

	port := strconv.Itoa(int(Ep.Port))
	err = AddHostKeyToKnownHosts(Ep.Host, port)
	if err != nil {
		t.Errorf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := CreateConfig()
	if err != nil {
		t.Errorf("Failed to create SSH config: %v", err)
	}

	conn, err := NewSSHConn(Ep, cfg)
	if err != nil {
		t.Errorf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Errorf("Failed to create SSH session: %v", err)
	}
	defer sess.Close()

	buf := &bytes.Buffer{}
	sess.Stdout = buf
	sess.Run("echo 'Hello, World!'")

	if buf.String() != "Hello, World!\n" {
		t.Errorf("Expected 'Hello, World!' got: %s", buf.String())
	}
}
