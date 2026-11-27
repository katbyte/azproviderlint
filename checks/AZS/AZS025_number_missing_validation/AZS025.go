// Package AZS025 defines an analyzer that reports a number property the user sets that has no validation.
package AZS025

import (
	"flag"
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

// strict is whether a number that is only checked for not being zero or negative is reported too.
var strict bool

func init() {
	Analyzer.Flags.Init("AZS025", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&strict, "strict", false, "also report a number whose only validation is that it is not zero, or at least 0 or 1: validation.IntAtLeast(0) or (1), FloatAtLeast(0) or (1), IntPositive, IntNotInSlice([]int{0}) or NoZeroValues, or the framework's AtLeast(0) or (1) and NoneOf(0)")
}

func run(pass *analysis.Pass) (any, error) {
	for _, value := range property.Values(pass) {
		// when strict, being told a number is not zero or negative does not count
		weak := strict && value.Loose
		if (value.Kind != property.Int && value.Kind != property.Float) || (value.Checked && !weak) {
			continue
		}

		what, wrong := "number property", "has no validation"
		if value.Element {
			what, wrong = "the numbers in this property", "have no validation"
		}
		if weak {
			wrong = "is only checked for not being zero or negative"
			if value.Element {
				wrong = "are only checked for not being zero or negative"
			}
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
