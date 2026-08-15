// Package AZR collects the resource implementation checks.
package AZR

import (
	"golang.org/x/tools/go/analysis"

	AZR001 "github.com/katbyte/azproviderlint/checks/AZR/AZR001_set_id_dereferenced_pointer"
	AZR002 "github.com/katbyte/azproviderlint/checks/AZR/AZR002_combined_create_update_method"
	AZR003 "github.com/katbyte/azproviderlint/checks/AZR/AZR003_resource_data_get_in_delete"
	AZR004 "github.com/katbyte/azproviderlint/checks/AZR/AZR004_resource_id_equality_comparison"
	AZR005 "github.com/katbyte/azproviderlint/checks/AZR/AZR005_case_insensitive_segments_feature_flag"
	AZR006 "github.com/katbyte/azproviderlint/checks/AZR/AZR006_stop_context_without_timeouts"
	AZR007 "github.com/katbyte/azproviderlint/checks/AZR/AZR007_state_change_conf"
)

// Checks contains all AZR (resource implementation) analyzers.
var Checks = []*analysis.Analyzer{
	AZR001.Analyzer,
	AZR002.Analyzer,
	AZR003.Analyzer,
	AZR004.Analyzer,
	AZR005.Analyzer,
	AZR006.Analyzer,
	AZR007.Analyzer,
}
