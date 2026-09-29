// Package AZS collects the schema & typed SDK model checks.
package AZS

import (
	"golang.org/x/tools/go/analysis"

	AZS001 "github.com/katbyte/azproviderlint/checks/AZS/AZS001_typed_sdk_model_64bit_types"
	AZS002 "github.com/katbyte/azproviderlint/checks/AZS/AZS002_schema_default_type_mismatch"
	AZS003 "github.com/katbyte/azproviderlint/checks/AZS/AZS003_schema_allows_empty_block"
	AZS004 "github.com/katbyte/azproviderlint/checks/AZS/AZS004_enum_validation_possible_values"
	AZS007 "github.com/katbyte/azproviderlint/checks/AZS/AZS007_optional_computed_missing_comment"
	AZS009 "github.com/katbyte/azproviderlint/checks/AZS/AZS009_computed_only_field_input_attributes"
)

// Checks contains all AZS (schema & typed SDK model) analyzers.
var Checks = []*analysis.Analyzer{
	AZS001.Analyzer,
	AZS002.Analyzer,
	AZS003.Analyzer,
	AZS004.Analyzer,
	AZS007.Analyzer,
	AZS009.Analyzer,
}
