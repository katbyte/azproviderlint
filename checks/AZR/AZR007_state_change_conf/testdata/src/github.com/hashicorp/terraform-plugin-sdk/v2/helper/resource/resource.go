// Package resource is a minimal stand-in for terraform-plugin-sdk/v2/helper/resource used only
// by the AZR007 analysistest fixtures. The real package aliases StateChangeConf from the retry
// package, which this fixture mirrors.
package resource

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

// StateChangeConf aliases retry.StateChangeConf, mirroring the real helper package.
type StateChangeConf = retry.StateChangeConf
