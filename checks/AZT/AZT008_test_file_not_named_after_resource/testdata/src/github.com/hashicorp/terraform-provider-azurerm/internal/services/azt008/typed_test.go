package azt008_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azt008"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azt008other"
)

// Should be flagged: neither typed_resource.go nor typed_data_source.go gives this file its name
func TestAccTyped_misplaced(t *testing.T) { // want `TestAccTyped_misplaced builds test data for resource "azurerm_typed", declared in typed_resource.go, so it belongs in typed_resource_test.go or typed_resource_<part>_test.go`
	data := acceptance.BuildTestData(t, azt008.TypedResource{}.ResourceType(), "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

func TestAccTypedDataSource_misplaced(t *testing.T) { // want `TestAccTypedDataSource_misplaced builds test data for data source "azurerm_typed", declared in typed_data_source.go, so it belongs in typed_data_source_test.go or typed_data_source_<part>_test.go`
	data := acceptance.BuildTestData(t, fmt.Sprintf("data.%s", azt008.TypedDataSource{}.ResourceType()), "test")
	data.DataSourceTest(t, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: a type of the same name from another package is not this package's
func TestAccTyped_otherPackage(t *testing.T) {
	data := acceptance.BuildTestData(t, azt008other.TypedResource{}.ResourceType(), "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}
