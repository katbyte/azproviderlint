package azt004

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

func BuildName(prefix string) string {
	return prefix + "-thing"
}

// Should NOT be flagged: go test only runs functions from _test.go files
func TestNotInATestFile(t *testing.T) {
	acceptance.BuildTestData(t, "azurerm_thing", "test")
}
