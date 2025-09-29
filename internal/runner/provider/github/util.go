package github

import (
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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
