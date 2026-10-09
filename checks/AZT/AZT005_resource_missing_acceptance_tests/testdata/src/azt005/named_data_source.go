package azt005

type NamedDataSource struct{}

func (r NamedDataSource) ResourceType() string {
	return "azurerm_named"
}
