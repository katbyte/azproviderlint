package azt005_test

import "testing"

// found by name: the older AzureRM prefix and a Resource suffix are both accepted
func TestAccAzureRMLegacy_basic(t *testing.T)           {}
func TestAccLegacyResource_requiresImport(t *testing.T) {}

// a data source may carry DataSource at either end
func TestAccDataSourceNamed_basic(t *testing.T) {}

// a different resource's tests never count for another
func TestAccTypedOther_list_update(t *testing.T) {}
