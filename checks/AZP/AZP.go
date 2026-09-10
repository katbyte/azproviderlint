// Package AZP collects the provider-wide convention checks.
package AZP

import (
	"golang.org/x/tools/go/analysis"

	AZP001 "github.com/katbyte/azproviderlint/checks/AZP/AZP001_microsoft_docs_url_locale"
)

// Checks contains all AZP (provider-wide convention) analyzers.
var Checks = []*analysis.Analyzer{
	AZP001.Analyzer,
}
