package github

import (
	"errors"
	"fmt"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/crypto/ssh"
)

// ClonePullRequest clones the pull request repository to the specified directory.
// this is used when an incoming pull request is received.
// if the repository is private, you need to specify a tokoen.
func (payload *PullRequestPayload) ClonePullRequest(token string, dir string) (*git.Repository, error) {
	branchRef := payload.PullRequest.OriginBranch.Ref
	cloneURL := payload.PullRequest.OriginBranch.Repo.CloneURL

	var auth *http.BasicAuth
	if token != "" {
		// Username can be anything except empty since Github ignores this field
		// "x-access-token" is conventional
		auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}
	logger.Printf("Cloning repository %v", payload.Repository.Name)
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		Auth:          auth,
		URL:           cloneURL,
		ReferenceName: plumbing.NewBranchReferenceName(branchRef),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		logger.Printf("Failed to clone URL: %v | branch: %v", cloneURL, branchRef)
		return nil, err
	}
	logger.Printf("Repository cloned successfully to directory: %v", dir)
	return repo, err
}

// Clone clones the repository into dir.
func (reader *GitRepoReader) Clone(conn *ssh.Client, dir string) (string, error) {
	logger.Printf("Cloning repository %v...", reader.CloneURL)
	cmd := fmt.Sprintf("mkdir -p %s && cd %s && git clone %s", dir, dir, reader.CloneURL)

	s, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer s.Close()

	b, err := s.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Fetch fetches the remote origin of the repository.
// it relies on the remote origin being set in the repository reader.
func (reader *GitRepoReader) Fetch(conn *ssh.Client, dir string) (string, error) {
	logger.Printf("Fetching repository %v...", reader.CloneURL)
	if reader.RemoteOrigin == "" {
		return "", errors.New("reader.RemoteOrigin is empty")
	}

	cmd := fmt.Sprintf("cd %s && git fetch origin %s", dir, reader.RemoteOrigin)
	s, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer s.Close()

	b, err := s.CombinedOutput(cmd)
	if err != nil {
		return string(b), fmt.Errorf("git fetch failed (cmd %q): %w — output:\n%s", cmd, err, string(b))
	}
	return string(b), nil
}

// CreateWorkTree creates a worktree in the repository.
// it uses the BranchName field in the reader and should typically set
// if using a pull request to the created branch after fetch.
func (reader *GitRepoReader) CreateWorkTree(conn *ssh.Client, repoDir, wrkTreeRelPath string) error {
	logger.Printf("Creating worktree %v...", filepath.Join(repoDir, wrkTreeRelPath))
	if reader.BranchName == "" {
		return errors.New("reader.BranchName is empty")
	}
	cmd := fmt.Sprintf("cd %s && git worktree add %s %s", repoDir, wrkTreeRelPath, reader.BranchName)
	s, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Run(cmd)

}

// RemoveWorkTree removes a worktree from the repository.
func (reader *GitRepoReader) RemoveWorkTree(conn *ssh.Client, repoDir, wrkTreeRelPath string) error {
	logger.Printf("Removing worktree %v...", filepath.Join(repoDir, wrkTreeRelPath))

	cmd := fmt.Sprintf("cd %s && git worktree remove %s", repoDir, wrkTreeRelPath)
	s, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Run(cmd)
}
