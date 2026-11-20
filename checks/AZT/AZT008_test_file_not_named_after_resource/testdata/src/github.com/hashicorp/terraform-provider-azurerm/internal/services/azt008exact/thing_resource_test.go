package azt008exact_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

type ThingResource struct{}

// Should NOT be flagged: exactly thing_resource_test.go
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
