package config

import (
	"testing"
)

func TestValidatePipeline(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "valid-pipeline",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							Pattern:  "pattern",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid-pipeline-with-files",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							File:     []string{"file1", "path/dir/file2"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid-pipeline-with-strategy-conflict",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							File:     []string{"file1", "path/dir/file2"},
							Pattern:  "pattern",
						},
					},
				},
			},
			wantErr: ErrFileStrategyConflict{TaskName: "task1"},
		},
		{
			name: "invalid-pipeline-no-strategy",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
						},
					},
				},
			},
			wantErr: ErrNoFileStrategySpecified,
		},
		{
			name: "invalid-pipeline-no-task-cmd",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:    "task1",
							RunsOn:  []string{"host1", "host2"},
							Pattern: "pattern",
						},
					},
				},
			},
			wantErr: ErrNoCmdsSpecified{TaskName: "task1"},
		},
		{
			name: "invalid-pipeline-no-runs-on",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							Pattern:  "pattern",
						},
					},
				},
			},
			wantErr: ErrNoTaskRunsOnSpecified{TaskName: "task1"},
		},
		{
			name: "invalid-pipeline-no-task-name",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							Pattern:  "pattern",
						},
					},
				},
			},
			wantErr: ErrNoTaskNameSpecified,
		},
		{
			name: "invalid-pipeline-no-tasks",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name:       "build-task",
						BuildSteps: []string{"step1", "step2"},
					},
				},
			},
			wantErr: ErrNoTasksSpecified,
		},
		{
			name: "invalid-pipeline-no-build-name",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						BuildSteps: []string{"step1", "step2"},
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							Pattern:  "pattern",
						},
					},
				},
			},
			wantErr: ErrEmptyBuildName,
		},
		{
			name: "invalid-pipeline-no-build-steps",
			cfg: Config{
				Pipeline: Pipeline{
					Build: BuildTaskProducer{
						Name: "build-task",
					},
					Tasks: []TaskConsumerJobs{
						{
							Name:     "task1",
							RunsOn:   []string{"host1", "host2"},
							Commands: []string{"cmd1", "cmd2", "cmd3 -v"},
							Pattern:  "pattern",
						},
					},
				},
			},
			wantErr: ErrEmptyBuildSteps,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksRunParallelSet := []bool{}
			for _, task := range tt.cfg.Pipeline.Tasks {
				if task.RunsInParallel == nil {
					tasksRunParallelSet = append(tasksRunParallelSet, true)
				} else {
					tasksRunParallelSet = append(tasksRunParallelSet, *task.RunsInParallel)
				}
			}

			err := tt.cfg.ValidatePipeline()
			if err != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, err)
			}

			for i, task := range tt.cfg.Pipeline.Tasks {
				if *task.RunsInParallel != tasksRunParallelSet[i] {
					t.Errorf("Expected task %v with RunsInParaller value of %v, got: %v",
						task.Name, *task.RunsInParallel, tasksRunParallelSet[i])
				}
			}
		})
	}
}
