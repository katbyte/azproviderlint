package azs005

import (
	"github.com/example/provider/pluginsdk"
	"github.com/example/provider/sdk"
)

type Registration struct {
	autoRegistration autoRegistration
}

// autoRegistration mirrors azurerm's generated auto-registration: the wrapper methods above
// delegate to these via an append spread, and the entries here must still be collected.
type autoRegistration struct{}

func (autoRegistration) Resources() []sdk.Resource {
	return []sdk.Resource{
		MissingAutoResource{}, // want `resource "azurerm_missing_auto" has no corresponding data source`
	}
}

func (autoRegistration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		CoveredAutoDataSource{},
	}
}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_covered_untyped": untypedResource(), // covered by SupportedDataSources
		"azurerm_cross_covered":   untypedResource(), // covered by the typed DataSources entry
		"azurerm_missing_untyped": untypedResource(), // want `resource "azurerm_missing_untyped" has no corresponding data source`
	}

	if featureFlag() {
		resources["azurerm_missing_conditional"] = untypedResource() // want `resource "azurerm_missing_conditional" has no corresponding data source`
	}

	// keys that do not resolve to a constant are skipped on the resource side: a local
	// variable, a call, and package vars without a constant string initializer
	localName := "azurerm_local"
	resources[localName] = untypedResource()
	resources[dynamicName()] = untypedResource()
	resources[computedName] = untypedResource()
	resources[uninitialisedName] = untypedResource()

	return resources
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_covered_untyped": untypedResource(),
	}
}

func (r Registration) Resources() []sdk.Resource {
	out := []sdk.Resource{
		CoveredTypedResource{},
		CoveredAutoResource{},     // covered by the auto-registered data source
		MissingTypedResource{},    // want `resource "azurerm_missing_typed" has no corresponding data source`
		MissingMultiVarResource{}, // want `resource "azurerm_multi_var" has no corresponding data source`
		RunCommandResource{},      // action-style resource — never reported
	}
	out = append(out, r.autoRegistration.Resources()...)

	if featureFlag() {
		out = append(out, MissingAppendedResource{}) // want `resource "azurerm_missing_appended" has no corresponding data source`
	}

	// a pointer element with a pointer-receiver ResourceType resolves like a value element
	out = append(out, &MissingPointerResource{}) // want `resource "azurerm_missing_pointer" has no corresponding data source`

	// an element whose static type has no ResourceType declaration in the package is skipped
	var extra sdk.Resource
	out = append(out, extra)

	return out
}

func (r Registration) DataSources() []sdk.DataSource {
	dataSources := []sdk.DataSource{
		CoveredTypedDataSource{},
		CrossCoveredDataSource{},
	}
	dataSources = append(dataSources, r.autoRegistration.DataSources()...)
	return dataSources
}

func (r Registration) FrameworkResources() []sdk.FrameworkWrappedResource {
	return []sdk.FrameworkWrappedResource{
		CoveredFrameworkResource{},
		MissingFrameworkResource{}, // want `resource "azurerm_missing_framework" has no corresponding data source`
	}
}

func (r Registration) FrameworkDataSources() []sdk.FrameworkWrappedDataSource {
	return []sdk.FrameworkWrappedDataSource{
		CoveredFrameworkDataSource{},
	}
}

// lookalike shares the registration method names but not their return shapes, so its
// entries are never collected.
type lookalike struct{}

func (lookalike) Resources() []string {
	return []string{"azurerm_lookalike"}
}

func (lookalike) SupportedResources() map[string]string {
	return map[string]string{"azurerm_lookalike": ""}
}

func untypedResource() *pluginsdk.Resource {
	return &pluginsdk.Resource{}
}

func dynamicName() string {
	return "azurerm_dynamic"
}

var (
	computedName      = dynamicName()
	uninitialisedName string
)

func featureFlag() bool {
	return false
}
