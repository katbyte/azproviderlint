// Package AZP collects the provider-wide convention checks.
package AZP

import (
	"golang.org/x/tools/go/analysis"

	AZP001 "github.com/katbyte/azproviderlint/checks/AZP/AZP001_microsoft_docs_url_locale"
	AZP002 "github.com/katbyte/azproviderlint/checks/AZP/AZP002_resource_missing_data_source"
	AZP003 "github.com/katbyte/azproviderlint/checks/AZP/AZP003_data_source_missing_properties"
	AZP004 "github.com/katbyte/azproviderlint/checks/AZP/AZP004_registration_entries_sorted"
	AZP005 "github.com/katbyte/azproviderlint/checks/AZP/AZP005_case_insensitive_segments_feature_flag"
)

// Checks contains all AZP (provider-wide convention) analyzers.
var Checks = []*analysis.Analyzer{
	AZP001.Analyzer,
	AZP002.Analyzer,
	AZP003.Analyzer,
	AZP004.Analyzer,
	AZP005.Analyzer,
}
