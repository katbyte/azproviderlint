package azr011

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azr011/stubs"
)

type AssertedResource struct{ stubs.Resource }

var _ sdk.Resource = AssertedResource{}

type AssertedPointerResource struct{ stubs.Resource }

var _ sdk.Resource = &AssertedPointerResource{}

type AssertedNilResource struct{ stubs.Resource }

var _ sdk.Resource = (*AssertedNilResource)(nil)

// any of the sdk's interfaces will do: which one a type should name is the author's call
type AssertedWiderResource struct{ stubs.ResourceWithUpdate }

var _ sdk.Resource = AssertedWiderResource{}

type (
	GroupedResource   struct{ stubs.Resource }
	GroupedDataSource struct{ stubs.DataSource }
)

var (
	_ sdk.Resource   = GroupedResource{}
	_ sdk.DataSource = GroupedDataSource{}
)

// the same methods under another name check the same thing
type FrameworkAssertedAction struct{}

var _ stubs.Invoker = &FrameworkAssertedAction{}

func (*FrameworkAssertedAction) Invoke() {}

// asserted in another file of the package
type ElsewhereResource struct{ stubs.Resource }

// the sdk has no use for these
type Model struct {
	Name string `tfschema:"name"`
}

type Named interface{ ResourceType() string }

type Generic[T any] struct{ stubs.Resource }
