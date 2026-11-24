// Package AZS011 defines an analyzer that reports resource and property descriptions with nothing in them.
package AZS011

import (
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
)

// Analyzer checks for descriptions that are empty or only spaces. A blank one reads as done and documents nothing, so it is worse than leaving
// the description out.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS011",
	Doc:      "check for resource and property descriptions that are blank",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS011_description_blank/README.md",
	Requires: []*analysis.Analyzer{description.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, found := range description.All(pass) {
		if strings.TrimSpace(found.Text) == "" {
			pass.Reportf(found.Value.Pos(), "description is blank - write one, or remove it")
		}
	}

	return nil, nil
}
