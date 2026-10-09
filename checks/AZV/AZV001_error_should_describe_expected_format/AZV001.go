// Package AZV001 defines an analyzer that reports unclear 'invalid format of' error messages
// that should describe the expected format instead.
package AZV001

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for error messages containing "invalid format of" which don't tell the
// user how to fix the problem, e.g. `invalid format of "foo"` should instead describe the
// expected format: `"foo" must start with a letter, may contain letters and numbers, ...`.
var Analyzer = &analysis.Analyzer{
	Name:     "AZV001",
	Doc:      "check for unclear 'invalid format of' error messages that should describe the expected format instead",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZV/AZV001_error_should_describe_expected_format/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return
		}

		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return
		}

		if strings.Contains(value, "invalid format of ") {
			pass.Reportf(lit.Pos(), "unclear error message: describe the expected format instead of 'invalid format of'")
		}
	})

	return nil, nil
}
