// Package AZT collects the acceptance testing checks.
package AZT

import (
	"golang.org/x/tools/go/analysis"

	AZT001 "github.com/katbyte/azproviderlint/checks/AZT/AZT001_acceptance_test_external_package"
	AZT002 "github.com/katbyte/azproviderlint/checks/AZT/AZT002_credentials_from_environment"
	AZT003 "github.com/katbyte/azproviderlint/checks/AZT/AZT003_misnamed_acceptance_test"
	AZT004 "github.com/katbyte/azproviderlint/checks/AZT/AZT004_acceptance_test_not_named_test_acc"
	AZT006 "github.com/katbyte/azproviderlint/checks/AZT/AZT006_update_test_single_config"
	AZT007 "github.com/katbyte/azproviderlint/checks/AZT/AZT007_stray_acc_in_test_name"
	AZT008 "github.com/katbyte/azproviderlint/checks/AZT/AZT008_test_file_not_named_after_resource"
)

// Checks contains all AZT (acceptance testing) analyzers.
var Checks = []*analysis.Analyzer{
	AZT001.Analyzer,
	AZT002.Analyzer,
	AZT003.Analyzer,
	AZT004.Analyzer,
	AZT006.Analyzer,
	AZT007.Analyzer,
	AZT008.Analyzer,
}
