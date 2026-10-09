package azt004

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should be flagged: in-package test files are checked too
func TestBuildName_live(t *testing.T) { // want `TestBuildName_live runs an acceptance test, name it TestAccBuildName_live so acceptance runs pick it up`
	acceptance.BuildTestData(t, "azurerm_thing", "test")
}

// Should NOT be flagged: a unit test
func TestBuildName(t *testing.T) {
	if BuildName("a") != "a-thing" {
		t.Fatal("wrong name")
	}
}
