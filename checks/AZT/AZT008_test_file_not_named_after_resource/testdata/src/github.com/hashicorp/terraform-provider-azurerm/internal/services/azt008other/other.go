// Package azt008other declares a type with the same name as one in the package under test.
package azt008other

type TypedResource struct{}

func (r TypedResource) ResourceType() string {
	return "azurerm_typed"
}
