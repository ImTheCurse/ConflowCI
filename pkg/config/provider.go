package config

import (
	"errors"
	"fmt"
)

var ErrEmptyPAToken = errors.New("Empty Personal Access Token")
var ErrEmptyBranch = errors.New("Empty branch name")
var ErrEmptyRepository = errors.New("Empty repository name")

func (cfg *Config) GetCloneURL() string {
	repo := cfg.Provider.Github.Repository
	cloneURL := fmt.Sprintf("https://github.com/%v.git", repo)

	if cfg.Provider.Github.Auth != nil {
		token := cfg.Provider.Github.Auth.Token
		cloneURL = fmt.Sprintf("https://ci:%v@github.com/%v.git", token, repo)
	}
	return cloneURL
}

func (provider *Provider) ValidateProvider() error {
	if len(provider.Github.Repository) <= 0 {
		return ErrEmptyRepository
	}
	if len(provider.Github.Branch) <= 0 {
		return ErrEmptyBranch
	}
	if provider.Github.Auth != nil {
		if len(provider.Github.Auth.Token) <= 0 {
			return ErrEmptyPAToken
		}
	}

	return nil
}
