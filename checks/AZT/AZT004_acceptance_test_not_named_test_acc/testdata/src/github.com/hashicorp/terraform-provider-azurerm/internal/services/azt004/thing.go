package azt004

import (
	"os"
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

// Configured is production code that reads TF_ACC to pick a default; calling it does not make
// a test an acceptance test
func Configured() bool {
	return os.Getenv("TF_ACC") != ""
}
