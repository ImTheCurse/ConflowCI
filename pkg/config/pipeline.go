package config

import (
	"errors"
	"fmt"
)

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

func (cfg *Config) ValidatePipeline() error {
	for i, task := range cfg.Pipeline.Tasks {
		// set RunInParallel to true by default
		if task.RunsInParallel == nil {
			cfg.Pipeline.Tasks[i].RunsInParallel = new(bool)
			*cfg.Pipeline.Tasks[i].RunsInParallel = true
		}
	}

	pipeline := cfg.Pipeline
	if pipeline.Build.Name == "" {
		return ErrEmptyBuildName
	}
	// Test build exist, as we use it for running
	if len(pipeline.Build.BuildSteps) == 0 {
		return ErrEmptyBuildSteps
	}

	tasks := pipeline.Tasks
	if len(tasks) == 0 {
		return ErrNoTasksSpecified
	}

	for _, task := range tasks {
		if task.Name == "" {
			return ErrNoTaskNameSpecified
		}
		if len(task.RunsOn) == 0 {
			return ErrNoTaskRunsOnSpecified{TaskName: task.Name}
		}

		if len(task.Commands) == 0 {
			return ErrNoCmdsSpecified{TaskName: task.Name}
		}

		pattern := task.Pattern
		files := task.File

		if pattern == "" && len(files) == 0 {
			return ErrNoFileStrategySpecified
		}
		if pattern != "" && len(files) > 0 {
			return ErrFileStrategyConflict{TaskName: task.Name}
		}
	}
	return nil
}
