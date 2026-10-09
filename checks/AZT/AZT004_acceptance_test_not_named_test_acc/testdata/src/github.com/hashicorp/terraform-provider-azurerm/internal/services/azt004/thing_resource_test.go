package azt004_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azt004"
)

type ThingResource struct{}

// Should NOT be flagged: named as an acceptance test
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: a unit test
func TestThing_validation(t *testing.T) {
	if azt004.BuildName("a") != "a-thing" {
		t.Fatal("wrong name")
	}
}

// Should be flagged: runs through the acceptance framework under a plain Test name
func TestThing_basic(t *testing.T) { // want `TestThing_basic runs an acceptance test, name it TestAccThing_basic so acceptance runs pick it up`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

func TestThingDataSource_basic(t *testing.T) { // want `TestThingDataSource_basic runs an acceptance test, name it TestAccThingDataSource_basic so acceptance runs pick it up`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.DataSourceTest(t, []acceptance.TestStep{{Config: "basic"}})
}

// Should be flagged: through the plugin testing harness directly
func TestThing_list(t *testing.T) { // want `TestThing_list runs an acceptance test, name it TestAccThing_list so acceptance runs pick it up`
	resource.Test(t, resource.TestCase{Steps: []resource.TestStep{{Config: "list"}}})
}

// Should be flagged: the acceptance test is in a helper, called or handed over as a value
func TestThing_viaHelper(t *testing.T) { // want `TestThing_viaHelper runs an acceptance test, name it TestAccThing_viaHelper so acceptance runs pick it up`
	testThing_helper(t)
}

func TestThing_sequential(t *testing.T) { // want `TestThing_sequential runs an acceptance test, name it TestAccThing_sequential so acceptance runs pick it up`
	acceptance.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic": testThing_helper,
		},
	})
}

// Should NOT be flagged: lowercase helpers are not run by go test
func testThing_helper(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "helper"}})
}

// Should be flagged: talks to Azure by hand behind a TF_ACC check
func TestThing_cliAuth(t *testing.T) { // want `TestThing_cliAuth runs an acceptance test, name it TestAccThing_cliAuth so acceptance runs pick it up`
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
}

// Should be flagged: Acc is in the name, just not where the runner looks
func TestWebAppAccActiveSlot_basic(t *testing.T) { // want `TestWebAppAccActiveSlot_basic runs an acceptance test, rename it to start with TestAcc so acceptance runs pick it up`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: a method is not a test function, so no runner selects it
func (r ThingResource) TestThing_receiver(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, r, []acceptance.TestStep{{Config: "method"}})
}
