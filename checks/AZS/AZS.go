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
	AZS010 "github.com/katbyte/azproviderlint/checks/AZS/AZS010_description_missing"
	AZS011 "github.com/katbyte/azproviderlint/checks/AZS/AZS011_description_blank"
	AZS012 "github.com/katbyte/azproviderlint/checks/AZS/AZS012_description_indefinite_article"
	AZS013 "github.com/katbyte/azproviderlint/checks/AZS/AZS013_description_acronym_case"
	AZS014 "github.com/katbyte/azproviderlint/checks/AZS/AZS014_description_repeats_schema"
	AZS015 "github.com/katbyte/azproviderlint/checks/AZS/AZS015_description_punctuation"
	AZS016 "github.com/katbyte/azproviderlint/checks/AZS/AZS016_description_product_name_case"
	AZS017 "github.com/katbyte/azproviderlint/checks/AZS/AZS017_description_opens_with_specifies"
	AZS018 "github.com/katbyte/azproviderlint/checks/AZS/AZS018_description_note"
	AZS019 "github.com/katbyte/azproviderlint/checks/AZS/AZS019_description_question"
)

// Checks contains all AZS (schema & typed SDK model) analyzers.
var Checks = []*analysis.Analyzer{
	AZS001.Analyzer,
	AZS002.Analyzer,
	AZS003.Analyzer,
	AZS004.Analyzer,
	AZS007.Analyzer,
	AZS009.Analyzer,
	AZS010.Analyzer,
	AZS011.Analyzer,
	AZS012.Analyzer,
	AZS013.Analyzer,
	AZS014.Analyzer,
	AZS015.Analyzer,
	AZS016.Analyzer,
	AZS017.Analyzer,
	AZS018.Analyzer,
	AZS019.Analyzer,
}
