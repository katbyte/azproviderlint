// Package AZT collects the acceptance testing checks.
package AZT

import (
	"golang.org/x/tools/go/analysis"

	AZT001 "github.com/katbyte/azproviderlint/checks/AZT/AZT001_acceptance_test_external_package"
	AZT002 "github.com/katbyte/azproviderlint/checks/AZT/AZT002_credentials_from_environment"
	AZT003 "github.com/katbyte/azproviderlint/checks/AZT/AZT003_misnamed_acceptance_test"
)

// Checks contains all AZT (acceptance testing) analyzers.
var Checks = []*analysis.Analyzer{
	AZT001.Analyzer,
	AZT002.Analyzer,
	AZT003.Analyzer,
}
