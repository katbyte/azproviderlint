// Package AZR collects the resource implementation checks.
package AZR

import (
	"golang.org/x/tools/go/analysis"

	AZR001 "github.com/katbyte/azproviderlint/checks/AZR/AZR001_set_id_dereferenced_pointer"
	AZR002 "github.com/katbyte/azproviderlint/checks/AZR/AZR002_combined_create_update_method"
	AZR003 "github.com/katbyte/azproviderlint/checks/AZR/AZR003_resource_data_get_in_delete"
	AZR004 "github.com/katbyte/azproviderlint/checks/AZR/AZR004_resource_id_equality_comparison"
	AZR006 "github.com/katbyte/azproviderlint/checks/AZR/AZR006_stop_context_without_timeouts"
	AZR007 "github.com/katbyte/azproviderlint/checks/AZR/AZR007_state_change_conf_custom_poller"
	AZR008 "github.com/katbyte/azproviderlint/checks/AZR/AZR008_flatten_returns_nil_slice"
	AZR009 "github.com/katbyte/azproviderlint/checks/AZR/AZR009_lifecycle_logging"
	AZR010 "github.com/katbyte/azproviderlint/checks/AZR/AZR010_flatten_handles_nil_input"
)

// Checks contains all AZR (resource implementation) analyzers.
var Checks = []*analysis.Analyzer{
	AZR001.Analyzer,
	AZR002.Analyzer,
	AZR003.Analyzer,
	AZR004.Analyzer,
	AZR006.Analyzer,
	AZR007.Analyzer,
	AZR008.Analyzer,
	AZR009.Analyzer,
	AZR010.Analyzer,
}
