// Package validation is a minimal stub of the plugin SDK's helper/validation, and of a provider's own wrapper of it, for the analyzer tests.
package validation

import "azs022/schema"

func StringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateFunc { return nil }

func StringInEnumSlice[T ~string](valid []T, ignoreCase bool) schema.SchemaValidateFunc { return nil }

func StringIsNotEmpty(any, string) ([]string, []error) { return nil, nil }
