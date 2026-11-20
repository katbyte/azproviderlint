package azt008

// the name is held in a constant
const typedName = "azurerm_typed"

type TypedDataSource struct{}

func (r TypedDataSource) ResourceType() string {
	return typedName
}
