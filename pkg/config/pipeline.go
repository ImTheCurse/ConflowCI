package config

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
