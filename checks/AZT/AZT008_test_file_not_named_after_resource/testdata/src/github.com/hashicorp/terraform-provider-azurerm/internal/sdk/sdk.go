// Package sdk is a minimal stub of the provider's typed SDK registration types.
package sdk

type Resource interface {
	ResourceType() string
}

type DataSource interface {
	ResourceType() string
}
