package azt006_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

type ThingResource struct{}

func (r ThingResource) basic(data acceptance.TestData) string    { return "basic" }
func (r ThingResource) complete(data acceptance.TestData) string { return "complete" }
func (r ThingResource) sku(data acceptance.TestData, sku string) string {
	return "sku " + sku
}

// Should NOT be flagged: two different configs
func TestAccThing_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.basic(data)},
		data.ImportStep(),
		{Config: r.complete(data)},
		data.ImportStep(),
	})
}

// Should NOT be flagged: the same helper with different arguments is a different config
func TestAccThing_updateSku(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.sku(data, "Basic")},
		{Config: r.sku(data, "Standard")},
	})
}

// Should NOT be flagged: not an update test
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.basic(data)},
	})
}

// Should NOT be flagged: update inside a word is a feature name, not the step
func TestAccThing_autoUpdateEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.basic(data)},
	})
}

// Should NOT be flagged: delegates to a helper, which is checked on its own
func TestAccThing_updateViaHelper(t *testing.T) {
	testAccThing_updateWith(t, "Standard")
}

func testAccThing_updateWith(t *testing.T, sku string) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.basic(data)},
		{Config: r.sku(data, sku)},
	})
}

// Should be flagged: a single config
func TestAccThing_updateTags(t *testing.T) { // want `TestAccThing_updateTags applies only one config, an update test needs a second, different config to apply`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.complete(data)},
		data.ImportStep(),
	})
}

// Should be flagged: the same config applied twice
func TestAccThing_list_update(t *testing.T) { // want `TestAccThing_list_update applies the same config every time, an update test needs a second, different config to apply`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.basic(data)},
		data.ImportStep(),
		{Config: r.basic(data)},
		data.ImportStep(),
	})
}

// Should be flagged: a sequential helper is an update test too
func testAccThing_updateIdentity(t *testing.T) { // want `testAccThing_updateIdentity applies only one config, an update test needs a second, different config to apply`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{Config: r.complete(data)},
	})
}
