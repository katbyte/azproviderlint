package azp003

import (
	"github.com/example/provider/fwschema"
	"github.com/example/provider/pluginsdk"
)

// typed pair: the resource's Arguments+Attributes expose "sku", the data source does not.

type TypedThingResource struct{}

func (r TypedThingResource) ResourceType() string {
	return "azurerm_typed_thing"
}

func (r TypedThingResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {Type: pluginsdk.TypeString, Required: true},
		"sku":  {Type: pluginsdk.TypeString, Required: true},
	}
}

func (r TypedThingResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"endpoint": {Type: pluginsdk.TypeString, Computed: true},
	}
}

type TypedThingDataSource struct{}

func (r TypedThingDataSource) ResourceType() string {
	return "azurerm_typed_thing"
}

func (r TypedThingDataSource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {Type: pluginsdk.TypeString, Required: true},
	}
}

func (r TypedThingDataSource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"endpoint": {Type: pluginsdk.TypeString, Computed: true},
	}
}

// FrameworkishResourceViaTyped exists so the typed slice has a second entry whose data source
// is absent; AZP003 must ignore it (AZP002's job).

type FrameworkishResourceViaTyped struct{}

func (r FrameworkishResourceViaTyped) ResourceType() string {
	return "azurerm_typed_no_ds"
}

func (r FrameworkishResourceViaTyped) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {Type: pluginsdk.TypeString, Required: true},
	}
}

// framework pair: the resource's Schema method declares "location", the data source's does
// not.

type FrameworkThingResource struct{}

func (r FrameworkThingResource) ResourceType() string {
	return "azurerm_framework_thing"
}

func (r FrameworkThingResource) Schema() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"name":     fwschema.StringAttribute{Required: true},
			"location": fwschema.StringAttribute{Required: true},
		},
		Blocks: map[string]fwschema.Block{},
	}
}

type FrameworkThingDataSource struct{}

func (r FrameworkThingDataSource) ResourceType() string {
	return "azurerm_framework_thing"
}

func (r FrameworkThingDataSource) Schema() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{Required: true},
		},
	}
}
