package azt008_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should NOT be flagged: the file is named after thing_data_source.go
func TestAccThingDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_thing", "test")
	data.DataSourceTest(t, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: run as a data source test under the bare resource name, so the data
// source's file is as right as the resource's
func TestAccThingDataSource_bareName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.DataSourceTest(t, []acceptance.TestStep{{Config: "basic"}})
}

// Should be flagged: a resource test in the data source's file
func TestAccThing_inDataSourceFile(t *testing.T) { // want `TestAccThing_inDataSourceFile builds test data for resource "azurerm_thing", declared in thing_resource.go, so it belongs in thing_resource_test.go or thing_resource_<part>_test.go`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
