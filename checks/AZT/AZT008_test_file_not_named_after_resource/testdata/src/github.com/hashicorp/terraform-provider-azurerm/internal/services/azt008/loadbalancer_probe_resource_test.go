package azt008_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
)

// Should be flagged once, on the first test: every test here is for the resource declared in
// lb_probe_resource.go, registered under a constant
func TestAccAzureRMLoadBalancerProbe_basic(t *testing.T) { // want `TestAccAzureRMLoadBalancerProbe_basic and 2 more in this file build test data for resource "azurerm_lb_probe", declared in lb_probe_resource.go, so they belong in lb_probe_resource_test.go or lb_probe_resource_<part>_test.go`
	data := acceptance.BuildTestData(t, "azurerm_lb_probe", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}})
}

func TestAccAzureRMLoadBalancerProbe_update(t *testing.T) {
	testAccAzureRMLoadBalancerProbe_protocol(t, "Http")
	data := acceptance.BuildTestData(t, "azurerm_lb_probe", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: "basic"}, {Config: "update"}})
}

// a helper that builds test data counts like any test
func testAccAzureRMLoadBalancerProbe_protocol(t *testing.T, protocol string) {
	data := acceptance.BuildTestData(t, "azurerm_lb_probe", "test")
	data.ResourceTest(t, ThingResource{}, []acceptance.TestStep{{Config: protocol}})
}
