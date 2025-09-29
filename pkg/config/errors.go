package config

import (
	"errors"
	"fmt"
)

type InvalidAddressFormat struct {
	address string
}

func (i InvalidAddressFormat) Error() string {
	return fmt.Sprintf("Invalid address format, expected ssh address format '[user@]address[:][port]' but got %s",
		i.address)
}

var ErrInvalidHost = errors.New("Empty host address")
var ErrInvalidPortNum = errors.New("Empty port number")
var ErrInvalidUser = errors.New("Empty username")
var ErrInvalidHostName = errors.New("Empty host name")
var ErrInvalidPrivateKeyPath = errors.New("Empty private key path")

var ErrEmptyBuildName = errors.New("Empty build name, atleast 1 character is required")
var ErrEmptyBuildSteps = errors.New("Build steps is empty, atleast 1 build step is required")
var ErrNoTasksSpecified = errors.New("No tasks specified, atleast 1 task is required")
var ErrNoTaskNameSpecified = errors.New("No task name specified, atleast 1 character is required")
var ErrNoFileStrategySpecified = errors.New("No file strategy found - specify files explictly or use a pattern to find files")

type ErrNoTaskRunsOnSpecified struct {
	TaskName string
}

func (e ErrNoTaskRunsOnSpecified) Error() string {
	return fmt.Sprintf("No machines specified to run on in task %s", e.TaskName)
}

type ErrNoCmdsSpecified struct {
	TaskName string
}

func (e ErrNoCmdsSpecified) Error() string {
	return fmt.Sprintf("No commands specified in task %s", e.TaskName)
}

type ErrFileStrategyConflict struct {
	TaskName string
}

func (e ErrFileStrategyConflict) Error() string {
	return fmt.Sprintf("File strategy conflict in task %s. either explictly specify files or a pattern.", e.TaskName)
}
