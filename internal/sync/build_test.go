package sync

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
)

func TestCreateMetadataFile(t *testing.T) {
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Errorf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")
	ctx := context.Background()
	container, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Errorf("Failed to start SSH server container: %v", err)
	}
	fmt.Println("SSH server running at", ssh.Ep.Host, ssh.Ep.Port)
	defer container.Terminate(ctx)

	port := strconv.Itoa(int(ssh.Ep.Port))
	err = ssh.AddHostKeyToKnownHosts(ssh.Ep.Host, port)
	if err != nil {
		t.Errorf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := ssh.CreateConfig()
	if err != nil {
		t.Errorf("Failed to create SSH config: %v", err)
	}

	conn, err := ssh.NewSSHConn(ssh.Ep, cfg)
	if err != nil {
		t.Errorf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	wb := &WorkerBuilder{
		Name:     "testproject",
		BuildID:  uuid.New(),
		CloneURL: "https://github.com/user/testproject.git",
		Conn:     conn,
	}

	if err := wb.CreateMetadataFile(); err != nil {
		t.Fatalf("CreateMetadataFile returned error: %v", err)
	}

	metadataPath := filepath.Join(buildPath, wb.Name, ".conflowci.toml")

	var metadata BuildMetadata

	s, err := conn.NewSession()
	if err != nil {
		t.Fatalf("Failed to create SSH session: %v", err)
	}
	defer s.Close()

	b, err := s.Output(fmt.Sprintf("cat %s", metadataPath))
	if err != nil {
		t.Fatalf("Failed to output metadata file: %v", err)
	}
	reader := bytes.NewReader(b)

	if err := toml.NewDecoder(reader).Decode(&metadata); err != nil {
		t.Fatalf("Failed to decode TOML: %v", err)
	}

	if metadata.Repository.Name != wb.Name {
		t.Errorf("Expected Repository.Name %s, got %s", wb.Name, metadata.Repository.Name)
	}
	if metadata.Repository.Source != wb.CloneURL {
		t.Errorf("Expected Repository.Source %s, got %s", wb.CloneURL, metadata.Repository.Source)
	}
	if metadata.State.Checksum == "" {
		t.Errorf("Expected non-empty checksum")
	}

	// Validate timestamps roughly (within 10 seconds)
	now := time.Now()
	clonedAt, _ := time.Parse(time.RFC3339, metadata.State.ClonedAt)
	if now.Sub(clonedAt) > 10*time.Second {
		t.Errorf("ClonedAt timestamp is too old")
	}
}
