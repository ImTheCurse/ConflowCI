package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	pb "github.com/ImTheCurse/ConflowCI/internal/provider/github/pb"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Clone clones the repository to the specified directory.
// this is usually called when an incoming event is received.
// if the repository is private, you need to specify a tokoen.
func (reader *GitRepoReader) Clone(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	branchName := req.BranchName
	cloneURL := req.CloneUrl

	var auth *http.BasicAuth
	if req.Token != "" {
		// Username can be anything except empty since Github ignores this field
		// "x-access-token" is conventional
		auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: req.Token,
		}
	}
	logger.Printf("Cloning repository %v", req.Name)
	_, err := git.PlainClone(req.Dir, false, &git.CloneOptions{
		Auth:          auth,
		URL:           cloneURL,
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return &pb.SyncResponse{Error: &pb.SyncError{
			Reason: fmt.Sprintf("Failed to clone URL: %v | branch: %v", cloneURL, branchName),
		},
			Output: "",
		}, nil
	}
	logger.Printf("Repository cloned successfully to directory: %v", req.Dir)
	return &pb.SyncResponse{
		Output: fmt.Sprintf("Repository cloned successfully to directory: %v", req.Dir),
		Error:  nil,
	}, nil
}

// Fetch fetches the remote origin of the repository.
// it relies on the remote origin being set in the repository reader.
func (reader *GitRepoReader) Fetch(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	repo, err := git.PlainOpen(req.Dir)
	if err != nil {
		return &pb.SyncResponse{Error: &pb.SyncError{Reason: fmt.Sprintf("Failed to open repository: %v", err)}}, nil
	}

	var auth *http.BasicAuth
	if req.Token != "" {
		// Username can be anything except empty since Github ignores this field
		// "x-access-token" is conventional
		auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: req.Token,
		}
	}

	remote, err := repo.Remote(req.RemoteOrigin)
	if err != nil {
		return &pb.SyncResponse{Error: &pb.SyncError{Reason: fmt.Sprintf("Failed to find remote origin: %v", err)}}, nil
	}
	logger.Printf("Fetching remote origin %v with spec: %v", req.RemoteOrigin, req.BranchRef)
	err = remote.Fetch(&git.FetchOptions{
		Auth: auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+%s", req.BranchRef)),
		},
		Progress: os.Stdout,
	})
	if err == git.NoErrAlreadyUpToDate {
		return &pb.SyncResponse{Output: "Repository already up to date"}, nil
	}
	if err != nil {
		return &pb.SyncResponse{Error: &pb.SyncError{Reason: fmt.Sprintf("Failed to fetch remote origin: %v", err)}}, nil
	}
	return &pb.SyncResponse{Output: "Repository fetched successfully", Error: nil}, nil
}

// CreateWorkTree creates a worktree in the repository.
func (reader *GitRepoReader) CreateWorkTree(ctx context.Context, req *pb.WorkTreeRequest) (*pb.SyncResponse, error) {
	args := []string{"worktree", "add", req.WorktreeRelPath, req.BranchName}
	cmd := exec.Command("git", args...)

	cmd.Dir = req.RepoDir
	b, err := cmd.CombinedOutput()

	if err != nil {
		return &pb.SyncResponse{Error: &pb.SyncError{
			Reason: fmt.Sprintf("Failed to create worktree: %v | output: %s", err, string(b)),
		}}, nil
	}
	logger.Println("Worktree created successfully")
	return &pb.SyncResponse{Output: "Worktree created successfully", Error: nil}, nil
}

// RemoveWorkTree removes a worktree from the repository.
func (reader *GitRepoReader) RemoveWorkTree(ctx context.Context, req *pb.WorkTreeRequest) (*pb.SyncResponse, error) {
	args := []string{"worktree", "remove", req.WorktreeRelPath}
	cmd := exec.Command("git", args...)

	cmd.Dir = req.RepoDir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return &pb.SyncResponse{Error: &pb.SyncError{
			Reason: fmt.Sprintf("Failed to remove worktree: %v | output: %s", err, string(b)),
		}}, nil
	}
	logger.Println("Worktree removed successfully")
	return &pb.SyncResponse{Output: "Worktree removed successfully", Error: nil}, nil
}
