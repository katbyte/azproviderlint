// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

import "azr013/cty"

type ResourceDiff struct{}

func (*ResourceDiff) GetRawConfig() cty.Value { return cty.Value{} }

type ResourceData struct{}

func (*ResourceData) GetRawConfig() cty.Value { return cty.Value{} }

type Schema struct {
	DiffSuppressFunc func(k, old, new string, d *ResourceData) bool
}
