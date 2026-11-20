package azt008exact_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should be flagged: with suffix off, a part before _test is no longer accepted
func TestAccThing_other(t *testing.T) { // want `TestAccThing_other builds test data for resource "azurerm_thing", declared in thing_resource.go, so it belongs in thing_resource_test.go$`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
