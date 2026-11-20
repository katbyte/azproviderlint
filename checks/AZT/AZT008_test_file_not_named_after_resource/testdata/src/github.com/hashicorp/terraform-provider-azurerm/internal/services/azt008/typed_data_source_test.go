package azt008_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azt008"
)

// Should NOT be flagged: the data source's ResourceType inside a Sprintf, in a file named after
// typed_data_source.go
func TestAccTypedDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, fmt.Sprintf("data.%s", azt008.TypedDataSource{}.ResourceType()), "test")
	data.DataSourceTest(t, []acceptance.TestStep{{Config: "basic"}})
}
