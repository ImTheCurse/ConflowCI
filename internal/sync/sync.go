package sync

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

var logger = log.New(os.Stdout, "[Sync]: ", log.Lshortfile|log.LstdFlags)

// Current state of a task.
type SyncState uint

const (
	StartingTask SyncState = iota
	RunningTask
	CompletedTask
	ErrorInTask
)

const buildPath string = "/home/conflowci/build"

// TaskDispatcher represents a task syncing for remote machines
// it tracks each state of the task, and is responsible for dispatching tasks
// to remote machines.
// It does the dispatching after the project is already built
type TaskDispatcher struct {
	TaskID uuid.UUID
	State  SyncState
	RunsOn []config.EndpointInfo
	Files  []string
}

func NewTaskDispatcher(conn *ssh.Client, cfg config.ValidatedConfig, task config.TaskConsumerJobs) (*TaskDispatcher, error) {
	files := []string{}
	var err error
	if task.File == nil {
		files, err = getFilesByRegex(conn, task.Pattern, buildPath)
		if err != nil {
			return nil, err
		}
	} else {
		files = task.File
	}
	return &TaskDispatcher{
		TaskID: uuid.New(),
		State:  StartingTask,
		RunsOn: getTasksMachine(cfg, task),
		Files:  files,
	}, err
}

func getTasksMachine(cfg config.ValidatedConfig, task config.TaskConsumerJobs) []config.EndpointInfo {
	res := []config.EndpointInfo{}
	for _, ep := range cfg.Endpoints {
		if slices.Contains(task.RunsOn, ep.Name) {
			res = append(res, ep)
		}
	}
	return res
}

func getFilesByRegex(conn *ssh.Client, expr string, buildDir string) ([]string, error) {
	s, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	cmd := fmt.Sprintf("find %s -regex %q", buildDir, expr)
	out, err := s.CombinedOutput(cmd)
	if err != nil {
		logger.Printf("Error running getFilesByRegex, output: %s", string(out))
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
