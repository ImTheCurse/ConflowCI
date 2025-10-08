package config

import (
	"errors"
	"fmt"
)

var ErrInvalidPersonalAccessToken = errors.New("Empty Personal Access Token")
var ErrInvalidBranchName = errors.New("Empty branch name")
var ErrInvalidRepoName = errors.New("Empty repository name")

// Gets the clone URL for the repository.
func (cfg *Config) GetCloneURL() string {
	repo := cfg.Provider.Github.Repository
	cloneURL := fmt.Sprintf("https://github.com/%v.git", repo)

	if cfg.Provider.Github.Auth != nil {
		token := cfg.Provider.Github.Auth.Token
		cloneURL = fmt.Sprintf("https://ci:%v@github.com/%v.git", token, repo)
	}
	return cloneURL
}

// Validates the configuration for the provider.
func (cfg *Config) ValidateProvider() error {
	if len(cfg.Provider.Github.Repository) <= 0 {
		return ErrInvalidRepoName
	}
	if len(cfg.Provider.Github.Branch) <= 0 {
		return ErrInvalidBranchName
	}
	if cfg.Provider.Github.Auth != nil {
		if len(cfg.Provider.Github.Auth.Token) <= 0 {
			return ErrInvalidPersonalAccessToken
		}
	}
	return nil
}
