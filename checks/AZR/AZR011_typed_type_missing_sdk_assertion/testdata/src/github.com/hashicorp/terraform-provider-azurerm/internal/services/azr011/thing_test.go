package azr011

import "github.com/hashicorp/terraform-provider-azurerm/internal/services/azr011/stubs"

// a type declared for a test is not the provider's
type TestOnlyResource struct{ stubs.Resource }
