package sync

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ImTheCurse/ConflowCI/internal/mq"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
	grpcUtil "github.com/ImTheCurse/ConflowCI/pkg/grpc"
)

func TestRunTaskOnAllMachines(t *testing.T) {
	grpcUtil.DefineFlags()
	*grpcUtil.TlsFlag = false
	flag.Parse()
	ctx := context.Background()
	logger.Printf("Creating Container RabbitMQ...")

	c, connURI, err := mq.CreateMessageQueueContainer()
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer c.Terminate(ctx)
	logger.Printf("Container RabbitMQ created")

	os.Setenv("CONFLOW_MQ_URI", connURI)
	defer os.Unsetenv("CONFLOW_MQ_URI")

	err = os.MkdirAll(os.ExpandEnv(BuildPath), 0777)
	if err != nil {
		t.Fatalf("Failed to create build path: %v", err)
	}
	defer os.RemoveAll(filepath.Join(os.ExpandEnv(BuildPath), "../"))
	f, err := os.Create(os.ExpandEnv(BuildPath) + "/test1.sh")
	if err != nil {
		t.Fatalf("Failed to create test1.sh: %v", err)
	}
	defer f.Close()
	if err != nil {
		t.Fatalf("Failed to create test1.sh: %v", err)
	}
	_, err = f.WriteString(`#!/bin/sh
		echo "hello-world!"`)
	if err != nil {
		t.Fatalf("Failed to write to test1.sh: %v", err)
	}
	err = f.Chmod(0777)
	if err != nil {
		t.Fatalf("Failed to chmod test1.sh: %v", err)
	}
	f.Close()
	ch := make(chan int)
	go mq.RunGRPCConsumerServer(ch)
	ep := config.EndpointInfo{
		Name: "test-1",
		Host: "localhost",
		Port: uint16(<-ch),
		User: "user",
	}

	valCfg := config.ValidatedConfig{
		Endpoints: []config.EndpointInfo{ep},
	}
	taskConsumer := config.TaskConsumerJobs{
		Name:     "task-runner-test",
		File:     []string{"test1.sh"},
		Commands: []string{"{file}"},
		RunsOn:   []string{"test-1"},
	}

	te, err := NewTaskExecutor(valCfg, taskConsumer, "")
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
	outputs := te.Outputs
	expected := []string{"hello-world!\n"}
	if !reflect.DeepEqual(outputs, expected) {
		t.Errorf("Expected outputs: %v, got: %v", expected, outputs)
	}
}
