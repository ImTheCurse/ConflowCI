package github

// PullRequestPayload represents the GitHub webhook payload for pull requests
type PullRequestPayload struct {
	Action      string      `json:"action"` // opened, closed, synchronize, etc.
	Number      int         `json:"number"` // PR number in the repo
	Sender      User        `json:"sender"` // Github user
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}

// PullRequest contains the details about the PR itself
type PullRequest struct {
	Title        string `json:"title"`
	User         User   `json:"user"` // who opened the PR
	OriginBranch Branch `json:"head"` // source branch
	TargetBranch Branch `json:"base"` // target branch
}

// Branch represents a branch in a repository
type Branch struct {
	Ref  string `json:"ref"`  // branch name
	SHA  string `json:"sha"`  // commit SHA of branch head
	Repo Repo   `json:"repo"` // repository the branch belongs to
}

// Repo represents a GitHub repository
type Repo struct {
	Name     string `json:"name"`
	CloneURL string `json:"clone_url"` // URL to clone repo
	Owner    User   `json:"owner"`     // repo owner
}

// Repository represents the target repository where PR is opened
type Repository struct {
	Name string `json:"name"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
	ID    int    `json:"id,omitempty"`
}
