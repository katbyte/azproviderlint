package azt008_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should NOT be flagged: a part between the declaring file's name and _test is fine
func TestAccThing_resourceIdentity(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
