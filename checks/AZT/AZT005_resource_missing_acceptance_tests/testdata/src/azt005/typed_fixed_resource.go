package azt005

import "azt005/sdk"

type TypedFixedResource struct{}

func (r TypedFixedResource) ResourceType() string {
	return "azurerm_typed_fixed"
}

func (r TypedFixedResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{}
}
