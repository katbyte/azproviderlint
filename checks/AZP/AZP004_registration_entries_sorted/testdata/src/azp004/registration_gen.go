package azp004

import "github.com/example/provider/sdk"

// autoRegistration mirrors azurerm's generated registration files: same shape, generated
// receiver name. An unsorted entry means the generator's input or template needs fixing.
type autoRegistration struct{}

func (autoRegistration) Resources() []sdk.Resource {
	return []sdk.Resource{
		ApiManagementResource{},
		WorkspaceResource{},
	}
}

func (autoRegistration) InvalidResources() []sdk.Resource {
	return []sdk.Resource{
		WorkspaceResource{},
		ApiManagementResource{}, // want `registration entries should be sorted alphabetically`
	}
}

func (autoRegistration) WebsiteCategories() []string {
	return []string{
		"Chaos Studio",
		"Compute",
	}
}
