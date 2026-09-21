package azs006

import (
	"github.com/example/provider/pluginsdk"
	"github.com/example/provider/sdk"
)

type Registration struct {
	autoRegistration autoRegistration
}

// autoRegistration mirrors azurerm's generated auto-registration: the wrapper methods
// delegate to these via an append spread, and the entries here are collected on their own.
type autoRegistration struct{}

func (autoRegistration) Resources() []sdk.Resource {
	return []sdk.Resource{
		AutoThingResource{},
	}
}

func (autoRegistration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		AutoThingDataSource{}, // want `data source "azurerm_auto_thing" is missing resource properties: "tier"`
	}
}

// registration keys held in package vars resolve through their constant initializer, even in
// a multi-name spec; a var without a constant initializer does not resolve and the entry is
// dropped.
var (
	unusedName, viaVarName = "azurerm_unused", "azurerm_via_var"
	dynamicName            = dynamicKey()
)

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_complete":     resourceComplete(),
		"azurerm_incomplete":   resourceIncomplete(),
		"azurerm_via_helper":   resourceViaHelper(),
		"azurerm_dynamic_keys": resourceDynamicKeys(),
		"azurerm_no_ds":        resourceNoDataSource(),
		"azurerm_secretive":    resourceSecretive(),
		viaVarName:             resourceViaVar(),
		"azurerm_via_method":   r.resourceViaMethod(),
		"azurerm_external":     pluginsdk.EmptyResource(),
		"azurerm_func_lit": func() *pluginsdk.Resource {
			return resourceIncomplete()
		}(),
		"azurerm_var_value": resourceVarValue,
	}
	resources[dynamicName] = resourceIncomplete()
	localName := "azurerm_local"
	resources[localName] = resourceIncomplete()

	if featureFlag() {
		resources["azurerm_conditional"] = resourceIncomplete()
	}

	return resources
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	dataSources := map[string]*pluginsdk.Resource{
		"azurerm_complete":     dataSourceComplete(),
		"azurerm_incomplete":   dataSourceIncomplete(), // want `data source "azurerm_incomplete" is missing resource properties: "backup.retention_days", "networking", "zone"`
		"azurerm_via_helper":   dataSourceViaHelper(),  // want `data source "azurerm_via_helper" is missing resource properties: "password"`
		"azurerm_dynamic_keys": dataSourceDynamicKeys(),
		"azurerm_secretive":    dataSourceSecretive(), // want `data source "azurerm_secretive" is missing resource properties: "primary_access_key"`
		"azurerm_ds_only":      dataSourceComplete(),
		viaVarName:             dataSourceIncomplete(), // want `data source "azurerm_via_var" is missing resource properties: "zone"`
		"azurerm_via_method":   dataSourceIncomplete(), // want `data source "azurerm_via_method" is missing resource properties: "zone"`
		// schemas that cannot be followed to a declaration in this package yield no
		// properties, and the pair is skipped
		"azurerm_external":  dataSourceIncomplete(),
		"azurerm_func_lit":  dataSourceIncomplete(),
		"azurerm_var_value": dataSourceIncomplete(),
	}

	if featureFlag() {
		dataSources["azurerm_conditional"] = dataSourceIncomplete() // want `data source "azurerm_conditional" is missing resource properties: "backup.retention_days", "networking", "zone"`
	}

	return dataSources
}

func (r Registration) Resources() []sdk.Resource {
	out := []sdk.Resource{
		TypedThingResource{},
		FrameworkishResourceViaTyped{},
	}
	out = append(out, r.autoRegistration.Resources()...)

	if featureFlag() {
		out = append(out, &PointerThingResource{})
	}

	// elements whose static type has no ResourceType declaration in the package are skipped:
	// an interface-typed variable, a plain constructor call, and a spread from a method that
	// is not a registration method
	var extra sdk.Resource
	out = append(out, extra)
	out = append(out, newTypedThing())
	out = append(out, r.extraResources()...)

	return out
}

func (r Registration) extraResources() []sdk.Resource {
	return nil
}

func newTypedThing() sdk.Resource {
	return TypedThingResource{}
}

func (r Registration) DataSources() []sdk.DataSource {
	out := []sdk.DataSource{
		TypedThingDataSource{}, // want `data source "azurerm_typed_thing" is missing resource properties: "sku"`
	}
	out = append(out, r.autoRegistration.DataSources()...)

	if featureFlag() {
		out = append(out, &PointerThingDataSource{}) // want `data source "azurerm_pointer_thing" is missing resource properties: "sku"`
	}

	return out
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

func (r Registration) resourceViaMethod() *pluginsdk.Resource {
	return resourceViaVar()
}

func resourceViaVar() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
			"zone": {Type: pluginsdk.TypeString, Optional: true},
		},
	}
}

var resourceVarValue = resourceIncomplete()

func featureFlag() bool {
	return false
}
