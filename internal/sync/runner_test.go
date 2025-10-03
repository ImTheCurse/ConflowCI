package sync

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ImTheCurse/ConflowCI/internal/mq"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
)

func TestRunTaskOnAllMachines(t *testing.T) {
	ctx := context.Background()
	logger.Printf("Creating Container RabbitMQ")
	c, connURI, err := mq.CreateMessageQueueContainer()

	os.Setenv("CONFLOW_MQ_URI", connURI)
	defer os.Unsetenv("CONFLOW_MQ_URI")

	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer c.Terminate(ctx)
	logger.Printf("Container RabbitMQ created")

	pub, _, err := crypto.GenerateKeys()
	if err != nil {
		t.Errorf("Failed to generate keys: %v", err)
	}
	defer os.RemoveAll("keys")

	///////////////////// First Connection //////////////////////////
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

	WriteFileCmd := `for i in $(seq 1 20); do echo '#!/bin/sh' > "test$i.sh"; echo "echo Checking test run $i" >> "test$i.sh"; done`
	changeFilePermissionsCmd := `for i in $(seq 1 20); do chmod +x "test$i.sh"; done`
	createFileCmd := `for i in $(seq 1 20); do touch "test$i.sh"; done`

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
	fmt.Println("Running all tasks...")
	err = te.RunTaskOnAllMachines()
	for _, cmdOutput := range te.Outputs {
		fmt.Printf("Command executed successfully, output: %s\n", cmdOutput)
	}
	if err != nil {
		t.Errorf("Failed to run task on all machines: got errors %v", err)
	}

	if len(te.Errors) > 0 {
		t.Errorf("Expected no errors, got errors: %v", te.Errors)
	}
	// order might seem funky but this is a lexographical order.
	outputs := te.Outputs
	expected := []string{
		"Checking test run 1",
		"Checking test run 10",
		"Checking test run 11",
		"Checking test run 12",
		"Checking test run 13",
		"Checking test run 14",
		"Checking test run 15",
		"Checking test run 16",
		"Checking test run 17",
		"Checking test run 18",
		"Checking test run 19",
		"Checking test run 2",
		"Checking test run 20",
		"Checking test run 3",
		"Checking test run 4",
		"Checking test run 5",
		"Checking test run 6",
		"Checking test run 7",
		"Checking test run 8",
		"Checking test run 9",
	}
	for i, o := range outputs {
		outputs[i] = strings.TrimSpace(o)
	}
	sort.Strings(outputs)
	if !reflect.DeepEqual(outputs, expected) {
		t.Errorf("Expected outputs: %v, got: %v", expected, outputs)
	}
}
