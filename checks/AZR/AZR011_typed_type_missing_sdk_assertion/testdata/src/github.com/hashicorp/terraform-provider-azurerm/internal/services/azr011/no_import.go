package azr011

import "github.com/hashicorp/terraform-provider-azurerm/internal/services/azr011/stubs"

// this file does not import the sdk, so the line is asked for but not written
type NoImportResource struct{ stubs.Resource } // want `NoImportResource has no "var _ sdk\.Resource = NoImportResource\{\}" line`
