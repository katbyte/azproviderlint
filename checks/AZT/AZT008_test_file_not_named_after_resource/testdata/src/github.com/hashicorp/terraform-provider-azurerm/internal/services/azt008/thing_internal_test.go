package azt008

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should be flagged: in-package test files are checked too
func TestAccThing_internal(t *testing.T) { // want `TestAccThing_internal builds test data for resource "azurerm_thing", declared in thing_resource.go, so it belongs in thing_resource_test.go or thing_resource_<part>_test.go`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, Registration{}, []acceptance.TestStep{{Config: "basic"}})
}
