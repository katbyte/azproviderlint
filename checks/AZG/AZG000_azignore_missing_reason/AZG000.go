// Package AZG000 defines an analyzer that reports '//azignore:' directives that do not
// carry a reason explaining why the check is suppressed.
package AZG000

import (
	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/checks/azignore"
)

// Analyzer checks that every '//azignore:<Rule>' directive includes a reason
// ('//azignore:AZR001,AZR003 - deliberate subset'), so each suppression documents why the
// flagged code is acceptable rather than silently accumulating.
var Analyzer = &analysis.Analyzer{
	Name: "AZG000",
	Doc:  "check that //azignore directives include a reason: '//azignore:<Rule> - <reason>'",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG000_azignore_missing_reason/README.md",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if _, reason, ok := azignore.ParseDirective(comment.Text); !ok || reason != "" {
					continue
				}

				pass.Reportf(comment.Pos(), "azignore directive must include a reason: '//azignore:<Rule> - <reason>'")
			}
		}
	}

	return nil, nil
}
