// Package pluginsdk is a minimal stand-in for the provider's internal/tf/pluginsdk helper
// package used only by the AZR007 analysistest fixtures. Just like azurerm's real wrapper,
// StateChangeConf here is a type alias for retry.StateChangeConf rather than a distinct type.
package pluginsdk

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// StateChangeConf aliases retry.StateChangeConf, mirroring azurerm's real pluginsdk wrapper.
type StateChangeConf = retry.StateChangeConf

type (
	RetryError = retry.RetryError
	RetryFunc  = retry.RetryFunc
)

// Retry wraps retry.Retry, mirroring azurerm's real pluginsdk wrapper. The wrapper itself is not a wait, so it is not flagged.
func Retry(timeout time.Duration, f RetryFunc) error {
	return retry.Retry(timeout, f)
}

// RetryableError wraps retry.RetryableError.
func RetryableError(err error) *RetryError {
	return retry.RetryableError(err)
}
