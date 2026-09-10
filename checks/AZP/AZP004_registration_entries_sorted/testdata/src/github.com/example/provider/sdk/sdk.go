// Package sdk is a minimal stand-in for azurerm's typed SDK registration types used only by the
// AZP004 analysistest fixtures.
package sdk

// Resource is a minimal stand-in for the typed SDK resource interface.
type Resource interface {
	ResourceType() string
}
