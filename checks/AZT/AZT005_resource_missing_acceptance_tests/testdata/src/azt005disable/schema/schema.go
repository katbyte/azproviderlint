// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ResourceFunc func() error

type Resource struct {
	Create ResourceFunc
	Read   ResourceFunc
	Update ResourceFunc
	Delete ResourceFunc
}

// External stands in for a resource declared in another package.
func External() *Resource {
	return &Resource{}
}
