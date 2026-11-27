// Package AZS025 defines an analyzer that reports a number property the user sets that has no validation.
package AZS025

import (
	"go/constant"
	"strconv"

	"github.com/katbyte/azproviderlint/lib/property"
	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a number the user sets is validated, so a value the API will refuse, such as one out of range, is caught when the plan is
// made and not part way through an apply.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS025",
	Doc:      "check that number properties the user sets have validation",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS025_number_missing_validation/README.md",
	Requires: []*analysis.Analyzer{property.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, value := range property.Values(pass) {
		if (value.Kind != property.Int && value.Kind != property.Float) || value.Checked {
			continue
		}

		what, wrong := "number property", "has no validation"
		if value.Element {
			what, wrong = "the numbers in this property", "have no validation"
		}
		if tv, isName := pass.TypesInfo.Types[value.Name]; isName && tv.Value != nil && tv.Value.Kind() == constant.String {
			what = "number property " + strconv.Quote(constant.StringVal(tv.Value))
			if value.Element {
				what = "the numbers in " + strconv.Quote(constant.StringVal(tv.Value))
			}
		}
		if value.From != nil {
			wrong += " in " + value.From.Pkg().Name() + "." + value.From.Name() + "()"
		}

		add := "a ValidateFunc that checks"
		if value.Framework {
			add = "Validators that check"
		}

		pass.Reportf(value.At.Pos(), "%s %s - add %s what the API accepts", what, wrong, add)
	}

	return nil, nil
}
