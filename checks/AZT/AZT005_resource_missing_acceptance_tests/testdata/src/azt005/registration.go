package azt005

import (
	"azt005/pluginsdk"
	"azt005/schema"
	"azt005/sdk"
)

type Registration struct{}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_thing":     resourceThing(),     // basic, requiresImport, complete, update all present
		"azurerm_fixed":     resourceFixed(),     // no Update: basic and requiresImport are enough
		"azurerm_nilupdate": resourceNilUpdate(), // Update: nil counts as no Update
		"azurerm_split":     resourceSplit(),     // tests split over several files named after the resource file
		"azurerm_legacy":    resourceLegacy(),    // tests in an unrelated file, found by name
		"azurerm_seq":       resourceSeq(),       // sequential: lowercase testAcc subtests
		"azurerm_external":  schema.External(),   // declared elsewhere: skipped
	}

	if featureFlag() {
		resources["azurerm_cond"] = resourceCond() // want `resource "azurerm_cond" is missing acceptance tests: TestAccCond_basic, TestAccCond_requiresImport`
	}

	return resources
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_thing":  dataSourceThing(),
		"azurerm_orphan": dataSourceOrphan(), // want `data source "azurerm_orphan" is missing acceptance tests: TestAccOrphan_basic`
	}
}

func (r Registration) Resources() []sdk.Resource {
	out := []sdk.Resource{
		TypedResource{},      // want `resource "azurerm_typed" is missing acceptance tests: TestAccTyped_update`
		TypedFixedResource{}, // want `resource "azurerm_typed_fixed" is missing acceptance tests: TestAccTypedFixed_requiresImport`
	}

	// an element whose static type has no declaration here is skipped
	var extra sdk.Resource
	out = append(out, extra)

	return out
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		NamedDataSource{},
	}
}

func (r Registration) FrameworkResources() []sdk.FrameworkWrappedResource {
	return []sdk.FrameworkWrappedResource{
		&PointerResource{}, // want `resource "azurerm_pointer" is missing acceptance tests: TestAccPointer_complete`
	}
}

func featureFlag() bool {
	return false
}

func noop() error {
	return nil
}
