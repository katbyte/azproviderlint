package azt013

import (
	"testing"

	"azt013/acceptance"
)

type StorageAccountResource struct{}

type StorageAcctTestResource struct{} // want `test resource StorageAcctTestResource is not named after azurerm_storage_account, which it tests - name it StorageAccountResource`

// the generated tests give each word one capital, so this is not the name they look for
type CdnFrontDoorRuleResource struct{} // want `test resource CdnFrontDoorRuleResource is not named after azurerm_cdn_frontdoor_rule, which it tests - name it CdnFrontdoorRuleResource`

// a struct shared by several resources can only be named after one of them
type LoadBalancerResource struct{}

type ThingDataSource struct{}

func TestAccStorageAccount_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{})
}

func TestAccStorageAccount_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_account", "test")

	data.ResourceSequentialTest(t, &StorageAcctTestResource{}, []acceptance.TestStep{})
}

// reported once, however many tests use it
func TestAccStorageAccount_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_account", "test")

	data.ResourceTest(t, StorageAcctTestResource{}, []acceptance.TestStep{})
}

func TestAccCdnFrontDoorRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cdn_frontdoor_rule", "test")

	data.ResourceTest(t, CdnFrontDoorRuleResource{}, []acceptance.TestStep{})
}

func TestAccLoadBalancer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_load_balancer", "test")

	data.ResourceTest(t, LoadBalancerResource{}, []acceptance.TestStep{})
}

func TestAccLoadBalancerRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_load_balancer_rule", "test")

	data.ResourceTest(t, LoadBalancerResource{}, []acceptance.TestStep{})
}

// a struct kept in another package is reported where it is handed over
func TestAccShared_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_shared_thing", "test")

	data.ResourceTest(t, acceptance.Shared{}, []acceptance.TestStep{}) // want `test resource Shared is not named after azurerm_shared_thing, which it tests - name it SharedThingResource`
}

// a data source's test hands no struct to what runs it, so its name is not checked
func TestAccThingDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_thing", "test")
	_ = ThingDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{})
}
