package azp003

import (
	"github.com/example/provider/pluginsdk"
	"github.com/example/provider/sdk"
)

type Registration struct{}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_complete":     resourceComplete(),
		"azurerm_incomplete":   resourceIncomplete(),
		"azurerm_via_helper":   resourceViaHelper(),
		"azurerm_dynamic_keys": resourceDynamicKeys(),
		"azurerm_no_ds":        resourceNoDataSource(),
		"azurerm_secretive":    resourceSecretive(),
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_complete":     dataSourceComplete(),
		"azurerm_incomplete":   dataSourceIncomplete(), // want `data source "azurerm_incomplete" is missing resource properties: "backup.retention_days", "networking", "zone"`
		"azurerm_via_helper":   dataSourceViaHelper(),
		"azurerm_dynamic_keys": dataSourceDynamicKeys(),
		"azurerm_secretive":    dataSourceSecretive(), // want `data source "azurerm_secretive" is missing resource properties: "primary_access_key"`
	}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		TypedThingResource{},
		FrameworkishResourceViaTyped{},
	}
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		TypedThingDataSource{}, // want `data source "azurerm_typed_thing" is missing resource properties: "sku"`
	}
}

func (r Registration) FrameworkResources() []sdk.FrameworkWrappedResource {
	return []sdk.FrameworkWrappedResource{
		FrameworkThingResource{},
	}
}

func (r Registration) FrameworkDataSources() []sdk.FrameworkWrappedDataSource {
	return []sdk.FrameworkWrappedDataSource{
		FrameworkThingDataSource{}, // want `data source "azurerm_framework_thing" is missing resource properties: "location"`
	}
}
