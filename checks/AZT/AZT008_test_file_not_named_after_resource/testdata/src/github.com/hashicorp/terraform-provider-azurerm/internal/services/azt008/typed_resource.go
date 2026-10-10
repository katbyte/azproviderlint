package azt008

type TypedResource struct{}

func (r TypedResource) ResourceType() string {
	return "azurerm_typed"
}
