package github

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestClonePullRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "clone_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name            string
		payload         *PullRequestPayload
		token           string
		dir             string
		CheckFilesExist []string
		expectError     bool
	}{
		{
			name: "clone with a valid public repo URL",
			payload: &PullRequestPayload{
				Repository: Repository{
					Name: "Hello-World",
				},
				PullRequest: PullRequest{
					OriginBranch: Branch{
						Ref: "test",
						SHA: "b3cbd5bbd7e81436d2eee04537ea2b4c0cad4cdf",
						Repo: Repo{
							Name:     "Hello-World",
							CloneURL: "https://github.com/octocat/Hello-World.git",
							Owner: User{
								Login: "octat",
							},
						},
					},
				},
			},
			token:           "",
			dir:             filepath.Join(tempDir, "test1"),
			expectError:     false,
			CheckFilesExist: []string{"CONTRIBUTING.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// make sure that dir dosen't exist, otherwise
			// clone will fail.
			os.RemoveAll(tt.dir)

			repo, err := tt.payload.ClonePullRequest(tt.token, tt.dir)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if repo == nil {
				t.Errorf("expected repository to exist, got nil.")
			}
			for _, fname := range tt.CheckFilesExist {
				fpath := filepath.Join(tt.dir, fname)
				if _, err := os.Stat(fpath); errors.Is(err, os.ErrNotExist) {
					t.Errorf("expected file %s to exist, got %v", fname, err)
				}
			}
		})
	}

}
