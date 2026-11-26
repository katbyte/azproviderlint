// Package validation is a minimal stub of the plugin SDK's helper/validation for the analyzer tests.
package validation

import "azs025/schema"

func IntAtLeast(min int) schema.SchemaValidateFunc { return nil }

func IntBetween(min, max int) schema.SchemaValidateFunc { return nil }
