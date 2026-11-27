// Package AZS024 defines an analyzer that reports a string property the user sets that has no validation.
package AZS024

import (
	"flag"
	"go/constant"
	"strconv"

	"github.com/katbyte/azproviderlint/lib/property"
	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a string the user sets is validated, so a value the API will refuse is caught when the plan is made and not part way
// through an apply.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS024",
	Doc:      "check that string properties the user sets have validation",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS024_string_missing_validation/README.md",
	Requires: []*analysis.Analyzer{property.Analyzer},
	Run:      run,
}

// strict is whether a string that is only checked for not being empty is reported too.
var strict bool

func init() {
	Analyzer.Flags.Init("AZS024", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&strict, "strict", false, "also report a string whose only validation is that it is not empty: validation.StringIsNotEmpty, StringIsNotWhiteSpace or NoZeroValues, or the framework's LengthAtLeast(1)")
}

func run(pass *analysis.Pass) (any, error) {
	for _, value := range property.Values(pass) {
		// when strict, being told a string is not empty does not count
		weak := strict && value.OnlyNotEmpty
		if value.Kind != property.String || (value.Checked && !weak) {
			continue
		}

		what := "string property"
		if value.Element {
			what = "the strings in this property"
		}
		if tv, isName := pass.TypesInfo.Types[value.Name]; isName && tv.Value != nil && tv.Value.Kind() == constant.String {
			what = "string property " + strconv.Quote(constant.StringVal(tv.Value))
			if value.Element {
				what = "the strings in " + strconv.Quote(constant.StringVal(tv.Value))
			}
		}

		wrong := "has no validation"
		switch {
		case weak && value.Element:
			wrong = "are only checked for not being empty"

		case weak:
			wrong = "is only checked for not being empty"

		case value.Element:
			wrong = "have no validation"
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
