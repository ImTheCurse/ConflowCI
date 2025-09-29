package ssh

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
)

var Ep config.EndpointInfo = config.EndpointInfo{
	Name:           "container-node",
	User:           "linuxserver.io",
	Host:           "localhost",
	Port:           2222,
	PrivateKeyPath: "keys/id_rsa",
}

func CreateConfig() (*ssh.ClientConfig, error) {
	sshCfg := SSHConnConfig{
		Username:       Ep.User,
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
	Ep.Port = uint16(mappedPort.Int())
	Ep.Host = host
	return sshContainer, nil
}

func AddHostKeyToKnownHosts(host string, port string) error {
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
