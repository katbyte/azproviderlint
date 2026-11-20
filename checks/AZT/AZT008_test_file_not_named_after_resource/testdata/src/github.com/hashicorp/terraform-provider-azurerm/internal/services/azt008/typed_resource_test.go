package azt008_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azt008"
)

// Should NOT be flagged: the type is named through the resource's own ResourceType method,
// and the file is named after typed_resource.go
func TestAccTyped_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, azt008.TypedResource{}.ResourceType(), "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

func TestAccTyped_pointer(t *testing.T) {
	data := acceptance.BuildTestData(t, (&azt008.TypedResource{}).ResourceType(), "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
