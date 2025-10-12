package sync

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/ImTheCurse/ConflowCI/internal/sync/pb"
	"github.com/ImTheCurse/ConflowCI/pkg/config"
)

func TestGetFilesByRegex(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "regex_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	args := []string{"example_test.go", "diff_test.go", "another_test.go"}

	cmd := exec.Command("touch", args...)
	cmd.Dir = tempDir

	o, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Failed to run command: %v. got output: %v", err, string(o))
	}

	expr := ".+_test.go"
	finder := pb.TaskFileFinder{
		Pattern:  expr,
		BuildDir: tempDir,
	}

	files, err := GetFilesByRegex(ctx, &finder)
	if err != nil {
		t.Errorf("Failed to get files by regex: %v", err)
	}

	expected := []string{
		filepath.Join(tempDir, "example_test.go"),
		filepath.Join(tempDir, "diff_test.go"),
		filepath.Join(tempDir, "another_test.go"),
	}
	slices.Sort(expected)
	slices.Sort(files.Files)
	if !reflect.DeepEqual(files.Files, expected) {
		t.Errorf("Expected files: %v got: %v", expected, files.Files)
	}
}

func TestGetTasksMachine(t *testing.T) {
	// Setup test data
	endpoints := []config.EndpointInfo{
		{
			Name:           "test-node-1",
			User:           "testuser",
			Host:           "192.168.1.101",
			Port:           8871,
			PrivateKeyPath: "/path",
		},
		{
			Name:           "test-node-2",
			User:           "testuser",
			Host:           "test.example.com",
			Port:           22,
			PrivateKeyPath: "/path",
		},
		{
			Name:           "test-node-3",
			User:           "testuser",
			Host:           "another.example.com",
			Port:           2222,
			PrivateKeyPath: "/path",
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
