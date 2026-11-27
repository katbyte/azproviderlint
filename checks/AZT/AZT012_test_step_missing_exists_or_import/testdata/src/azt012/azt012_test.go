package azt012

import (
	"testing"

	"azt012/acceptance"
)

type ThingResource struct{}

func (ThingResource) basic(data acceptance.TestData) string { return "" }

func (ThingResource) complete(data acceptance.TestData) string { return "" }

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

func TestAccNoChecks_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{ // want `step 1 of TestAccNoChecks_basic never checks the resource exists - add check.That\(data.ResourceName\).ExistsInAzure\(r\) to its Check` `step 1 of TestAccNoChecks_basic is not followed by an import step - add data.ImportStep\(\) after it`
			Config: r.basic(data),
		},
	})
}

// every step that applies a config is held to this, not only the first or the basic test's
func TestAccThing_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
		data.ImportStep(),
		{ // want `step 3 of TestAccThing_update is not followed by an import step`
			Config: r.complete(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
		{ // want `step 4 of TestAccThing_update never checks the resource exists`
			Config: r.basic(data),
		},
		data.ImportStep(),
		data.ApplyStep(r.complete, r), // want `step 6 of TestAccThing_update is not followed by an import step`
	})
}

// checking a value is not checking the resource is there
func TestAccOtherChecks_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceSequentialTest(t, r, []acceptance.TestStep{
		{ // want `step 1 of TestAccOtherChecks_basic never checks the resource exists`
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.CheckThat(data.ResourceName).Key("name").HasValue("thing"),
			),
		},
		data.ImportStep(),
	})
}

// trying to create the resource again, to show the provider refuses to, follows the step that made it
func TestAccThing_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
		data.RequiresImportErrorStep(r.basic),
	})
}

func TestAccThing_requiresImportWrittenOut(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).ExistsInAzure(r),
		},
		{
			Config:      r.basic(data),
			ExpectError: acceptance.RequiresImportError("azurerm_thing"),
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

// a step that expects an error, only plans, destroys, or checks the resource is gone has nothing to exist or import
func TestAccThing_noResource(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	r := ThingResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config:      r.basic(data),
			ExpectError: nil,
		},
		{
			Config:   r.basic(data),
			PlanOnly: true,
		},
		{
			Config:  r.basic(data),
			Destroy: true,
		},
		{
			Config: r.basic(data),
			Check:  acceptance.CheckThat(data.ResourceName).DoesNotExistInAzure(r),
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

// steps built somewhere else cannot be read
func TestAccBuilt_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")

	data.ResourceTest(t, ThingResource{}, steps(data))
}

func steps(data acceptance.TestData) []acceptance.TestStep { return nil }
