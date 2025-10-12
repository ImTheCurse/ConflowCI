package github

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ImTheCurse/ConflowCI/internal/provider/github/pb"
	"google.golang.org/protobuf/proto"
)

func handleProtoError(t *testing.T, err error, resp *pb.SyncResponse) {
	if resp.Error != nil {
		t.Errorf("Unexpected error: %s", resp.Error.Reason)
	}
	if err != nil {
		t.Errorf("Unexpected proto server error: %v", err)
	}
	if err != nil || resp.Error != nil {
		t.FailNow()
	}
}

func TestClone(t *testing.T) {
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
		branchName      string
		CheckFilesExist []string
		expectError     bool
	}{
		{
			name:       "clone with a valid public repo URL",
			branchName: "test",
			payload: &PullRequestPayload{
				Repository: Repository{
					Name: "Hello-World",
				},
				PullRequest: PullRequest{
					OriginBranch: Branch{
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

			branchRef := tt.payload.PullRequest.OriginBranch
			cloneURL := tt.payload.PullRequest.OriginBranch.Repo.CloneURL

			reader := GitRepoReader{}

			ctx := context.Background()
			syncReq := pb.SyncRequest{
				Name:       tt.payload.Repository.Name,
				BranchName: tt.branchName,
				BranchRef:  branchRef.Ref,
				CloneUrl:   cloneURL,
				Dir:        tt.dir,
			}
			resp, err := reader.Clone(ctx, &syncReq)
			handleProtoError(t, err, resp)
			// repo, err := tt.payload.ClonePullRequest(tt.token, tt.dir)

			if tt.expectError {
				if resp.Error == nil {
					t.Errorf("expected error but got nil")
				}
			}
			if resp.Error != nil {
				t.Errorf("Unexpected error: %v", err)
				return
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

func TestFetch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fetch_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	reader := GitRepoReader{}
	req := pb.SyncRequest{
		Name:         "demo-repo",
		CloneUrl:     "https://github.com/ImTheCurse/demo-repo",
		BranchRef:    "pull/6/head:pr-6",
		BranchName:   "pr-6",
		RemoteOrigin: "origin",
		Dir:          tempDir,
	}
	cloneReq, ok := proto.Clone(&req).(*pb.SyncRequest)
	if !ok {
		t.Fatal("failed to clone request")
	}
	cloneReq.BranchName = "main"
	resp, err := reader.Clone(ctx, cloneReq)
	handleProtoError(t, err, resp)
	resp, err = reader.Fetch(ctx, &req)
	handleProtoError(t, err, resp)

}

func TestWorkTree(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fetch_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	reader := GitRepoReader{}
	fetchReq := pb.SyncRequest{
		Name:         "demo-repo",
		CloneUrl:     "https://github.com/ImTheCurse/demo-repo",
		BranchRef:    "pull/6/head:pr-6",
		BranchName:   "pr-6",
		RemoteOrigin: "origin",
		Dir:          tempDir,
	}
	cloneReq, ok := proto.Clone(&fetchReq).(*pb.SyncRequest)
	if !ok {
		t.Fatal("failed to clone request")
	}
	cloneReq.BranchName = "main"
	req := pb.WorkTreeRequest{
		Name:            cloneReq.Name,
		RepoDir:         tempDir,
		BranchName:      fetchReq.BranchName,
		WorktreeRelPath: "../tempWrkDir",
	}
	resp, err := reader.Clone(ctx, cloneReq)
	handleProtoError(t, err, resp)
	resp, err = reader.Fetch(ctx, &fetchReq)
	handleProtoError(t, err, resp)
	resp, err = reader.CreateWorkTree(ctx, &req)
	handleProtoError(t, err, resp)
	resp, err = reader.RemoveWorkTree(ctx, &req)
	handleProtoError(t, err, resp)
}
