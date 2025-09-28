package sync

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/ImTheCurse/ConflowCI/pkg/crypto"
	"github.com/ImTheCurse/ConflowCI/pkg/ssh"
)

func TestGetFilesByRegex(t *testing.T) {
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

	conn, err := ssh.NewSSHConn(Ep, cfg)
	if err != nil {
		t.Errorf("Failed to create SSH connection: %v", err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Errorf("Failed to create SSH session: %v", err)
	}
	defer sess.Close()

	cmd := "mkdir test && cd test && touch example_test.go diff_test.go another_test.go"
	o, err := sess.CombinedOutput(cmd)
	if err != nil {
		t.Errorf("Failed to run command: %v. got output: %v", err, string(o))
	}

	expr := ".+_test.go"

	files, err := getFilesByRegex(conn, expr, ".")
	if err != nil {
		t.Errorf("Failed to get files by regex: %v", err)
	}

	expected := []string{
		"./test/example_test.go",
		"./test/diff_test.go",
		"./test/another_test.go",
	}
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("Expected files: %v got: %v", expected, files)
	}
}

func TestGetTasksMachine(t *testing.T) {
	// Setup test data
	endpoints := []config.EndpointInfo{
		{
			Name: "test-node-1",
			User: "testuser",
			Host: "192.168.1.101",
			Port: 8871,
		},
		{
			Name: "test-node-2",
			User: "testuser",
			Host: "test.example.com",
			Port: 22,
		},
		{
			Name: "test-node-3",
			User: "testuser",
			Host: "another.example.com",
			Port: 2222,
		},
	}

	cfg := config.ValidatedConfig{
		Endpoints: endpoints,
	}

	tests := []struct {
		name     string
		task     config.TaskConsumerJobs
		expected []config.EndpointInfo
	}{
		{
			name: "single-matching-endpoint",
			task: config.TaskConsumerJobs{
				Name:   "test-task-1",
				RunsOn: []string{"test-node-1"},
			},
			expected: []config.EndpointInfo{endpoints[0]},
		},
		{
			name: "multiple-matching-endpoints",
			task: config.TaskConsumerJobs{
				Name:   "test-task-2",
				RunsOn: []string{"test-node-1", "test-node-3"},
			},
			expected: []config.EndpointInfo{endpoints[0], endpoints[2]},
		},
		{
			name: "all-matching-endpoints",
			task: config.TaskConsumerJobs{
				Name:   "test-task-3",
				RunsOn: []string{"test-node-1", "test-node-2", "test-node-3"},
			},
			expected: endpoints,
		},
		{
			name: "no-matching-endpoints",
			task: config.TaskConsumerJobs{
				Name:   "test-task-4",
				RunsOn: []string{"non-existent-node"},
			},
			expected: []config.EndpointInfo{},
		},
		{
			name: "partial-matching-endpoints",
			task: config.TaskConsumerJobs{
				Name:   "test-task-5",
				RunsOn: []string{"test-node-2", "non-existent-node", "test-node-3"},
			},
			expected: []config.EndpointInfo{endpoints[1], endpoints[2]},
		},
		{
			name: "empty-runs-on",
			task: config.TaskConsumerJobs{
				Name:   "test-task-6",
				RunsOn: []string{},
			},
			expected: []config.EndpointInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTasksMachine(cfg, tt.task)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d endpoints, got %d", len(tt.expected), len(result))
				return
			}

			for _, expectedEp := range tt.expected {
				if !slices.Contains(result, expectedEp) {
					t.Errorf("Expected endpoint: %v to be in result endpoints: %v", expectedEp, result)
				}
			}
		})
	}
}
