// Package typed is a minimal stand-in for a typed SDK resource package registered from another
// package, used only by the AZP004 analysistest fixtures.
package typed

// ComputeResource is a qualified typed SDK resource fixture.
type ComputeResource struct{}

// ResourceType implements sdk.Resource.
func (ComputeResource) ResourceType() string { return "azurerm_compute" }

// NetworkResource is a qualified typed SDK resource fixture.
type NetworkResource struct{}

// ResourceType implements sdk.Resource.
func (NetworkResource) ResourceType() string { return "azurerm_network" }
