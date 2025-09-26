package ssh

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
)

var ep config.EndpointInfo = config.EndpointInfo{
	User: "linuxserver.io",
	Host: "localhost",
	Port: 2222,
}

func CreateConfig() (*ssh.ClientConfig, error) {
	sshCfg := SSHConnConfig{
		Username:       ep.User,
		Password:       "testpass",
		PrivateKeyPath: "keys/id_rsa",
	}
	cfg, err := sshCfg.BuildConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to build SSH config: %v", err)
	}
	return cfg, nil
}

func CreateSSHServerContainer(pubKey string) (testcontainers.Container, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "lscr.io/linuxserver/openssh-server:latest",
		ExposedPorts: []string{"2222/tcp"},
		Env: map[string]string{
			"PUID":            "1000",
			"PGID":            "1000",
			"PASSWORD_ACCESS": "false",
			"PUBLIC_KEY":      pubKey,
		},

		WaitingFor: wait.ForListeningPort("2222/tcp").WithStartupTimeout(30 * time.Second),
	}
	sshContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}
	host, err := sshContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}
	mappedPort, err := sshContainer.MappedPort(ctx, "2222")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %v", err)
	}
	ep.Port = uint16(mappedPort.Int())
	ep.Host = host
	return sshContainer, nil
}

func addHostKeyToKnownHosts(host string, port string) error {
	cmd := exec.Command("ssh-keyscan", "-p", port, host)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	// Append to known_hosts
	f, err := os.OpenFile(fmt.Sprintf("%s/.ssh/known_hosts", userHomeDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(out.Bytes())
	return err
}

func TestNewSSHConn(t *testing.T) {
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Errorf("Failed to generate keys: %v", err)
	}
	ctx := context.Background()
	container, err := CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Errorf("Failed to start SSH server container: %v", err)
	}
	fmt.Println("SSH server running at", ep.Host, ep.Port)
	defer container.Terminate(ctx)

	port := strconv.Itoa(int(ep.Port))
	err = addHostKeyToKnownHosts(ep.Host, port)
	if err != nil {
		t.Errorf("Failed to add host key to known hosts: %v", err)
	}

	cfg, err := CreateConfig()
	if err != nil {
		t.Errorf("Failed to create SSH config: %v", err)
	}

	conn, err := NewSSHConn(ep, cfg)
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
	os.RemoveAll("keys")
}
