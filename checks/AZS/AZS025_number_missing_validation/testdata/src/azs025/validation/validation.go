// Package validation is a minimal stub of the plugin SDK's helper/validation for the analyzer tests.
package validation

import "azs025/schema"

func StringIsNotEmpty(any, string) ([]string, []error) { return nil, nil }

func StringIsNotWhiteSpace(any, string) ([]string, []error) { return nil, nil }

func NoZeroValues(any, string) ([]string, []error) { return nil, nil }

func StringLenBetween(low, high int) schema.SchemaValidateFunc { return nil }

func IntBetween(low, high int) schema.SchemaValidateFunc { return nil }

func FloatAtLeast(low float64) schema.SchemaValidateFunc { return nil }

func IntAtLeast(low int) schema.SchemaValidateFunc { return nil }

func IntPositive(any, string) ([]string, []error) { return nil, nil }

func IntNotInSlice(invalid []int) schema.SchemaValidateFunc { return nil }
