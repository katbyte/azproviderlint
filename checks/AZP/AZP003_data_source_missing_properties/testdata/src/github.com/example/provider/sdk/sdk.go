// Package sdk is a minimal stand-in for azurerm's typed and framework SDK registration types
// used only by the AZP003 analysistest fixtures.
package sdk

// Resource is a minimal stand-in for the typed SDK resource interface.
type Resource interface {
	ResourceType() string
}

// DataSource is a minimal stand-in for the typed SDK data source interface.
type DataSource interface {
	ResourceType() string
}

// FrameworkWrappedResource is a minimal stand-in for the framework wrapped resource interface.
type FrameworkWrappedResource interface {
	ResourceType() string
}

// FrameworkWrappedDataSource is a minimal stand-in for the framework wrapped data source interface.
type FrameworkWrappedDataSource interface {
	ResourceType() string
}
