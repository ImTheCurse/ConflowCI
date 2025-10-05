package provider

import "github.com/ImTheCurse/ConflowCI/pkg/config"

type RepositoryReader interface {
	Clone(targetEndpoint config.EndpointInfo, cloneURL, dir string) error
	Fetch(targetEndpoint config.EndpointInfo) error
	CreateWorkTree(targetEndpoint config.EndpointInfo, dir string) error
	RemoveWorkTree(targetEndpoint config.EndpointInfo, dir string) error
}
