package azt003_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azt003"
)

type ThingResource struct{}

// Should NOT be flagged: runs through the acceptance framework
func TestAccThing_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

func TestAccThingDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.DataSourceTest(t, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: runs through the plugin testing harness directly
func TestAccThing_list(t *testing.T) {
	resource.Test(t, resource.TestCase{Steps: []resource.TestStep{{Config: "list"}}})
}

func TestAccThing_writeOnly(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{Steps: []resource.TestStep{{Config: "writeOnly"}}})
}

// Should NOT be flagged: the acceptance test is in a helper
func TestAccThing_listen(t *testing.T) {
	testAccThing_rights(t, true, false)
}

func TestAccThing_send(t *testing.T) {
	testAccThing_rights(t, false, true)
}

func testAccThing_rights(t *testing.T, listen, send bool) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "rights"}})
}

// Should NOT be flagged: the helper is a method
func TestAccThing_method(t *testing.T) {
	ThingResource{}.run(t)
}

func (r ThingResource) run(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, r, []acceptance.TestStep{{Config: "method"}})
}

// Should NOT be flagged: sequential tests hand the helpers over as values
func TestAccThing_sequential(t *testing.T) {
	acceptance.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic": testAccThing_sequentialBasic,
		},
	})
}

func TestAccThing_sequentialByHand(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic": testAccThing_sequentialBasic,
		},
	}

	for group, m := range testCases {
		for name, tc := range m {
			t.Run(group, func(t *testing.T) {
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			})
		}
	}
}

func TestAccThing_subtests(t *testing.T) {
	t.Run("basic", testAccThing_sequentialBasic)
}

func testAccThing_sequentialBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_thing", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

// Should NOT be flagged: helpers that only reach the framework through each other
func TestAccThing_mutual(t *testing.T) {
	testAccThing_ping(t, 2)
}

func testAccThing_ping(t *testing.T, n int) {
	if n > 0 {
		testAccThing_pong(t, n-1)
	}
}

func testAccThing_pong(t *testing.T, n int) {
	testAccThing_ping(t, n)
	testAccThing_sequentialBasic(t)
}

// Should NOT be flagged: talks to Azure by hand behind a TF_ACC check
func TestAccThing_cliAuth(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set")
	}
}

func TestAccThing_envConstant(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip("TF_ACC not set")
	}
}

// Should NOT be flagged: Acc is the start of a word, not the acceptance prefix
func TestAccountName(t *testing.T) {
	if azt003.BuildName("a") != "a-thing" {
		t.Fatal("wrong name")
	}
}

// Should NOT be flagged: not a test function
func (r ThingResource) TestAccThing_receiver(t *testing.T) {}

// Should be flagged: unit tests named as acceptance tests
func TestAccThing_validation(t *testing.T) { // want `TestAccThing_validation does not run an acceptance test, name it TestThing_validation so it is not picked up as one`
	cases := []struct {
		Value    string
		Expected string
	}{
		{Value: "a", Expected: "a-thing"},
	}

	for _, tc := range cases {
		if got := azt003.BuildName(tc.Value); got != tc.Expected {
			t.Fatalf("expected %q, got %q", tc.Expected, got)
		}
	}
}

func TestAccThing_subtestsUnit(t *testing.T) { // want `TestAccThing_subtestsUnit does not run an acceptance test, name it TestThing_subtestsUnit so it is not picked up as one`
	t.Run("name", testThing_name)
}

func testThing_name(t *testing.T) {
	if azt003.BuildName("a") != "a-thing" {
		t.Fatal("wrong name")
	}
}

// Should be flagged: uses the framework without handing it the test
func TestAccThing_randomName(t *testing.T) { // want `TestAccThing_randomName does not run an acceptance test, name it TestThing_randomName so it is not picked up as one`
	if azt003.BuildName(acceptance.RandString(5)) == "" {
		t.Fatal("empty name")
	}
}
