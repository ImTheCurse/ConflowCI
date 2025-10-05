package sync

import (
	"log"
	"os"

	"github.com/ImTheCurse/ConflowCI/pkg/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
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

// Current state of a task.
type BuildState uint

const (
	StartingBuild BuildState = iota
	RunningBuild
	CompletedBuild
	ErrorInBuild
	CompleteBuildWithErrors
)

func (s BuildState) String() string {
	switch s {
	case StartingBuild:
		return "Starting build"
	case RunningBuild:
		return "Running build"
	case CompletedBuild:
		return "Completed build with no errors"
	case ErrorInBuild:
		return "Error building project"
	case CompleteBuildWithErrors:
		return "Completed build with errors"
	default:
		return "Unkown build state"
	}
}

const buildPath string = "$HOME/conflowci/build"
const metadataFileName string = ".conflowci.toml"

// TaskExecutor represents a task syncing for remote machines
// it tracks each state of the task, and is responsible for dispatching tasks
// to remote machines.
// It does the dispatching after the project is already built
type TaskExecutor struct {
	TaskID  uuid.UUID
	State   TaskState
	RunsOn  []config.EndpointInfo
	Files   []string
	Cmds    []string
	Outputs []string
	Errors  []string
}

type WorkerBuilder struct {
	Name     string
	BuildID  uuid.UUID
	State    BuildState
	RunsOn   []config.EndpointInfo
	Steps    []string
	CloneURL string
	Conn     *ssh.Client
}

type BuildMetadata struct {
	Repository RepositoryMetadata `toml:"Repository"`
	State      StateMetadata      `toml:"state"`
}
type StateMetadata struct {
	ClonedAt  string `toml:"cloned_at"`
	LastBuild string `toml:"last_build"`
	Checksum  string `toml:"checksum"`
}
type RepositoryMetadata struct {
	Name    string  `toml:"name"`
	Source  string  `toml:"source"`
	Version float32 `toml:"project_version"`
}
