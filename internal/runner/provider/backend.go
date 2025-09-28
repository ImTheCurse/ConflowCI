package provider

type ProviderOperator interface {
	Fetch() error
	CreateWorkTree() error
	RemoveWorkTree() error
	Build() error
}
