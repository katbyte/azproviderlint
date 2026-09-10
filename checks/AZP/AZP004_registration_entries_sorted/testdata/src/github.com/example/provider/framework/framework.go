// Package framework is a minimal stand-in for azurerm's plugin framework registration types
// used only by the AZP004 analysistest fixtures.
package framework

// Resource is a minimal stand-in for the framework resource interface.
type Resource interface {
	Metadata()
}
