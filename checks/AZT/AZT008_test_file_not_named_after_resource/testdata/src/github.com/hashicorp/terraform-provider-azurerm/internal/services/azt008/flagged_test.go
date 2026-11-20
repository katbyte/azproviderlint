package azt008_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should be flagged: the resource is registered behind a feature flag, and _resource is
// missing from this file's name
func TestAccFlagged_basic(t *testing.T) { // want `TestAccFlagged_basic builds test data for resource "azurerm_flagged", declared in flagged_resource.go, so it belongs in flagged_resource_test.go or flagged_resource_<part>_test.go`
	data := acceptance.BuildTestData(t, "azurerm_flagged", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
