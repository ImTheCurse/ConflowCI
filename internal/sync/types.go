package sync

import (
	"log"
	"os"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/google/uuid"
)

var logger = log.New(os.Stdout, "[Sync]: ", log.Lshortfile|log.LstdFlags)

// Current state of a task.
type TaskState uint

const (
	StartingTask TaskState = iota
	RunningTask
	CompletedTask
	ErrorInTask
	CompleteTaskWithErrors
)

func (s TaskState) String() string {
	switch s {
	case StartingTask:
		return "Starting task"
	case RunningTask:
		return "Running task"
	case CompletedTask:
		return "Completed task with no errors"
	case ErrorInTask:
		return "Error executing task"
	case CompleteTaskWithErrors:
		return "Completed with errors"
	default:
		return "Unkown task state"
	}
}

const buildPath string = "$HOME/conflowci/build"

// TaskExecutor represents a task syncing for remote machines
// it tracks each state of the task, and is responsible for dispatching tasks
// to remote machines.
// It does the dispatching after the project is already built
type TaskExecutor struct {
	TaskID      uuid.UUID
	State       TaskState
	RunsOn      []config.EndpointInfo
	Files       []string
	CmdQueue    chan string // TODO: change this into a message queue.
	OutputQueue chan string // TODO: change this to a message queue.
	ErrorQueue  chan error  // TODO: change this into a message queue.
}
