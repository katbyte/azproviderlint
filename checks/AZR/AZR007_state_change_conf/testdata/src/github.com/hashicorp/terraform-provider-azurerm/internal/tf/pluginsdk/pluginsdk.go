// Package pluginsdk is a minimal stand-in for the provider's internal/tf/pluginsdk helper
// package used only by the AZR007 analysistest fixtures. Just like azurerm's real wrapper,
// StateChangeConf here is a type alias for retry.StateChangeConf rather than a distinct type.
package pluginsdk

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

// StateChangeConf aliases retry.StateChangeConf, mirroring azurerm's real pluginsdk wrapper.
type StateChangeConf = retry.StateChangeConf
