// Package sdk is a minimal stand-in for azurerm's typed and framework SDK types.
package sdk

type ResourceFunc struct{}

type Resource interface {
	ResourceType() string
}

type DataSource interface {
	ResourceType() string
}

type FrameworkWrappedResource interface {
	ResourceType() string
}
