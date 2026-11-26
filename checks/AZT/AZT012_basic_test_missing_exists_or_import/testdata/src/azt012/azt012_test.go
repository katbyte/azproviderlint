package azt012

import (
	"testing"

	"azt012/acceptance"
)

type ThingResource struct{}

func (ThingResource) basic(data acceptance.TestData) string { return "" }

func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNoChecks_basic(t *testing.T) { // want `TestAccNoChecks_basic never checks the resource exists - add check.That\(data.ResourceName\).ExistsInAzure\(r\) to a step` `TestAccNoChecks_basic never imports the resource - add data.ImportStep\(\) after the config is applied`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
	})
}

// checking a value is not checking the resource is there, and failing to import one that exists is not importing it
func TestAccOtherChecks_basic(t *testing.T) { // want `TestAccOtherChecks_basic never checks the resource exists` `TestAccOtherChecks_basic never imports the resource`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.CheckThat(data.ResourceName).Key("name").HasValue("thing"),
			),
		},
		data.RequiresImportErrorStep(r.basic),
	})
}

func TestAccNoImport_basic(t *testing.T) { // want `TestAccNoImport_basic never imports the resource`
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
	})
}

// the other ways of importing
func TestAccImportBlock_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
		data.ImportBlockWithIDStep(false),
	})
}

func TestAccImportState_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
		{
			Config:      r.basic(data),
			ImportState: true,
		},
	})
}

// a data source has nothing to exist or import
func TestAccThingData_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_thing", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: ThingResource{}.basic(data),
		},
	})
}

// a data source tested with the resource's runner is still a data source, by its name or by what its data is built for
func TestAccThingDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
	})
}

func TestAccThingLookup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
	})
}

// a step that applies a config and checks the resource exists in one
func TestAccApplyStep_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.ApplyStep(r.basic, r),
		data.ImportStep(),
	})
}

// only the basic test is held to this, and steps built somewhere else cannot be read
func TestAccThing_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
		},
	})
}

func TestAccBuilt_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")

	data.ResourceTest(t, ThingResource{}, steps(data))
}

func steps(data acceptance.TestData) []acceptance.TestStep { return nil }
