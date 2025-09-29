package sync

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
)

func TestRunTaskOnAllMachines(t *testing.T) {
	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Errorf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")

	///////////////////// First Connection //////////////////////////
	ctx := context.Background()
	container, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Errorf("Failed to start SSH server container: %v", err)
	}
	Ep := ssh.Ep
	fmt.Println("SSH server running at", Ep.Host, Ep.Port)
	defer container.Terminate(ctx)

	port := strconv.Itoa(int(Ep.Port))
	err = ssh.AddHostKeyToKnownHosts(Ep.Host, port)
	if err != nil {
		t.Errorf("Failed to add host key to known hosts: %v", err)
	}
	cfg, err := ssh.CreateConfig()
	if err != nil {
		t.Errorf("Failed to create SSH config: %v", err)
	}
	conn1, err := ssh.NewSSHConn(Ep, cfg)
	if err != nil {
		t.Errorf("Failed to create SSH connection: %v", err)
	}
	defer conn1.Close()

	s1, err := conn1.NewSession()
	if err != nil {
		t.Errorf("Failed to create SSH session: %v", err)
	}
	defer s1.Close()

	path := "$HOME/conflowci/build"
	// createFileCmd := "touch test{1..20}.sh"
	// WriteFileCmd := `for i in {1..20}; do echo 'Checking test run ${i}' > "test${i}.sh"; done`
	WriteFileCmd := `for i in $(seq 1 20); do echo '#!/bin/sh' > "test$i.sh"; echo "echo Checking test run $i" >> "test$i.sh"; done`
	changeFilePermissionsCmd := `for i in $(seq 1 20); do chmod +x "test$i.sh"; done`

	createFileCmd := `for i in $(seq 1 20); do touch "test$i.sh"; done`
	// WriteFileCmd := `for i in $(seq 1 20); do echo "Checking test run $i" > "test$i.sh"; done`
	// changeFilePerrmissionsCmd := "chmod +x test{1..20}.sh"

	cmd := fmt.Sprintf("mkdir -p %s && cd %s && %s && %s && %s", path, path, createFileCmd, WriteFileCmd, changeFilePermissionsCmd)
	_, err = s1.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to execute command: %v", err)
	}
	///////////////////////////////// End First Connection /////////////////////////////////////////
	//////////////////////////////// Second Connection /////////////////////////////////////////////

	container2, err := ssh.CreateSSHServerContainer(string(pub))
	if err != nil {
		t.Errorf("Failed to start SSH server container(second): %v", err)
	}
	Ep2 := ssh.Ep
	Ep2.Name = "test-2"
	fmt.Println("Second SSH server running at", Ep2.Host, Ep2.Port)
	defer container2.Terminate(ctx)

	port2 := strconv.Itoa(int(Ep2.Port))
	err = ssh.AddHostKeyToKnownHosts(Ep2.Host, port2)
	if err != nil {
		t.Errorf("Failed to add host key to known hosts: %v", err)
	}
	cfg, err = ssh.CreateConfig()
	if err != nil {
		t.Errorf("Failed to create SSH config: %v", err)
	}
	conn2, err := ssh.NewSSHConn(Ep2, cfg)
	if err != nil {
		t.Errorf("Failed to create SSH connection: %v", err)
	}
	defer conn2.Close()

	s2, err := conn2.NewSession()
	if err != nil {
		t.Errorf("Failed to create SSH session: %v", err)
	}
	defer s2.Close()
	_, err = s2.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to execute command: %v", err)
	}
	/////////////////////////////// End Second Connection /////////////////////////////////////////

	pattern := ".*/test[0-9]{1,2}.sh"
	valCfg := config.ValidatedConfig{
		Endpoints: []config.EndpointInfo{Ep, Ep2},
	}
	taskConsumer := config.TaskConsumerJobs{
		Name:     "task-runner-test",
		Pattern:  pattern,
		Commands: []string{"{file}"},
		RunsOn:   []string{"container-node", "test-2"},
	}
	te, err := NewTaskExecutor(conn1, valCfg, taskConsumer)
	if err != nil {
		t.Errorf("Failed to create task executor: %v", err)
	}
	errors := te.RunTaskOnAllMachines()
	for cmdOutput := range te.OutputQueue {
		fmt.Printf("Command executed successfully, output: %s\n", cmdOutput)
	}
	if len(errors) > 0 {
		t.Errorf("Failed to run task on all machines: got errors %v", errors)
	}

}
