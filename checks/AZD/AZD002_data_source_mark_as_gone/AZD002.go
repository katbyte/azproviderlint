// Package AZD002 defines an analyzer that reports data sources using MarkAsGone instead of
// returning an error when the resource cannot be found.
package AZD002

import (
	"go/ast"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `metadata.MarkAsGone(...)` in data source files. Data Sources should
// return an error when a resource cannot be found rather than marking the resource as gone.
var Analyzer = &analysis.Analyzer{
	Name:     "AZD002",
	Doc:      "check for data sources using MarkAsGone instead of returning an error when the resource cannot be found",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZD/AZD002_data_source_mark_as_gone/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.SelectorExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		if !tf.InDataSourceFile(pass, n) {
			return
		}

		sel, ok := n.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "MarkAsGone" {
			return
		}

		pass.Reportf(sel.Pos(), "data sources should return an error when a resource cannot be found instead of calling MarkAsGone")
	})

	return nil, nil
}
