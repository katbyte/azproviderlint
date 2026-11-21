// Package validation is a minimal stub of the plugin SDK's helper/validation for the analyzer tests.
package validation

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func StringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateFunc { return nil }

func IntBetween(lowest, highest int) schema.SchemaValidateFunc { return nil }

func StringIsNotEmpty(any, string) ([]string, []error) { return nil, nil }
