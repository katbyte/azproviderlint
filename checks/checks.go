// Package checks exposes all azproviderlint analyzers, grouped by category.
package checks

import (
	"slices"

	"github.com/katbyte/azproviderlint/checks/AZC"
	"github.com/katbyte/azproviderlint/checks/AZD"
	"github.com/katbyte/azproviderlint/checks/AZG"
	AZG000 "github.com/katbyte/azproviderlint/checks/AZG/AZG000_azignore_missing_reason"
	"github.com/katbyte/azproviderlint/checks/AZP"
	"github.com/katbyte/azproviderlint/checks/AZR"
	"github.com/katbyte/azproviderlint/checks/AZS"
	"github.com/katbyte/azproviderlint/checks/AZT"
	"github.com/katbyte/azproviderlint/checks/AZV"
	"github.com/katbyte/azproviderlint/checks/azignore"
)

// All contains every azproviderlint analyzer across all categories, each honouring
// '//azignore:<Name> - <reason>' comment directives — except AZG000, which polices the
// directives themselves and is deliberately unwrapped (and kept out of AZG.Checks) so a
// bare directive cannot suppress the report about itself by listing AZG000.
var All = append(azignore.Wrap(slices.Concat(
	AZC.Checks,
	AZD.Checks,
	AZG.Checks,
	AZP.Checks,
	AZR.Checks,
	AZS.Checks,
	AZT.Checks,
	AZV.Checks,
)), AZG000.Analyzer)
