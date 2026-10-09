package azt005

import "azt005/sdk"

type TypedResource struct{}

func (r TypedResource) ResourceType() string {
	return "azurerm_typed"
}

func (r TypedResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{}
}

func (r TypedResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{}
}
