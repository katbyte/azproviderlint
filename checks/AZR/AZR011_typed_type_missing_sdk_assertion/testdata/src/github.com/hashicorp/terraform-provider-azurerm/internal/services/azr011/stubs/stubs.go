// Package stubs holds the methods of each kind of sdk type, for the fixtures to embed rather than repeat.
package stubs

import "github.com/hashicorp/terraform-provider-azurerm/internal/sdk"

type DataSource struct{}

var _ sdk.DataSource = DataSource{}

func (DataSource) Arguments() map[string]string  { return nil }
func (DataSource) Attributes() map[string]string { return nil }
func (DataSource) ModelObject() any              { return nil }
func (DataSource) ResourceType() string          { return "" }
func (DataSource) Read() sdk.ResourceFunc        { return sdk.ResourceFunc{} }

type Resource struct{ DataSource }

var _ sdk.Resource = Resource{}

func (Resource) Create() sdk.ResourceFunc             { return sdk.ResourceFunc{} }
func (Resource) Delete() sdk.ResourceFunc             { return sdk.ResourceFunc{} }
func (Resource) IDValidationFunc() func(string) error { return nil }

type ResourceWithUpdate struct{ Resource }

var _ sdk.ResourceWithUpdate = ResourceWithUpdate{}

func (ResourceWithUpdate) Update() sdk.ResourceFunc { return sdk.ResourceFunc{} }

type FrameworkResource struct{}

var _ sdk.FrameworkWrappedResourceWithUpdate = FrameworkResource{}

func (FrameworkResource) ModelObject() any            { return nil }
func (FrameworkResource) ResourceType() string        { return "" }
func (FrameworkResource) Create(sdk.ResourceMetadata) {}
func (FrameworkResource) Read(sdk.ResourceMetadata)   {}
func (FrameworkResource) Update(sdk.ResourceMetadata) {}
func (FrameworkResource) Delete(sdk.ResourceMetadata) {}

// Invoker has the methods of sdk.Action under another name, as the framework's own interface does.
type Invoker interface {
	Invoke()
}
