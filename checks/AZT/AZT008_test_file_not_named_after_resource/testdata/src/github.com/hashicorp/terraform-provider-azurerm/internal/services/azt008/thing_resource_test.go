package azt008_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

type ThingResource struct{}

// Should NOT be flagged: the file is named after thing_resource.go
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: also builds test data for the data source, and the file is right for
// the resource
func TestAccThing_withDataSource(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	dataSource := acceptance.BuildTestData(t, "data.azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: dataSource.ResourceName}})
}

// Should NOT be flagged: nothing is registered under these names in this package
func TestAccThing_unregistered(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_elsewhere", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

func TestAccThing_ephemeral(t *testing.T) {
	data := acceptance.BuildTestData(t, "ephemeral.azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: the type is not known until the test runs
func TestAccThing_variable(t *testing.T) {
	for _, name := range []string{"azurerm_lb_probe", "azurerm_flagged"} {
		data := acceptance.BuildTestData(t, name, "test")
		data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
	}
}

// Should be flagged: a test for another resource has strayed into this file
func TestAccThing_probe(t *testing.T) { // want `TestAccThing_probe builds test data for resource "azurerm_lb_probe", declared in lb_probe_resource.go, so it belongs in lb_probe_resource_test.go or lb_probe_resource_<part>_test.go`
	data := acceptance.BuildTestData(t, "azurerm_lb_probe", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
