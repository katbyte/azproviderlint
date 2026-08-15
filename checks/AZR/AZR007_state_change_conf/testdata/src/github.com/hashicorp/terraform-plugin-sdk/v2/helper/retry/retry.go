// Package retry is a minimal stand-in for terraform-plugin-sdk/v2/helper/retry used only by
// the AZR007 analysistest fixtures. It declares the canonical StateChangeConf type that the
// provider wrapper packages alias.
package retry

import (
	"context"
	"time"
)

// StateChangeConf is a minimal stand-in for retry.StateChangeConf.
type StateChangeConf struct {
	Pending []string
	Target  []string
	Timeout time.Duration
}

// WaitForStateContext is a stub matching the real helper's signature.
func (c *StateChangeConf) WaitForStateContext(ctx context.Context) (interface{}, error) {
	return nil, nil
}
