// Package AZD001 defines an analyzer that reports data sources calling SetId with an empty
// string instead of returning an error when the resource cannot be found.
package AZD001

import (
	"go/ast"
	"go/token"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `d.SetId("")` in data source files. Data Sources should return an
// error when a resource cannot be found rather than setting an empty ID.
var Analyzer = &analysis.Analyzer{
	Name:     "AZD001",
	Doc:      "check for data sources calling SetId with an empty string instead of returning an error when the resource cannot be found",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZD/AZD001_data_source_empty_set_id/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		if !tf.InDataSourceFile(pass, n) {
			return
		}

		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "SetId" {
			return
		}

		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || lit.Value != `""` {
			return
		}

		pass.Reportf(call.Pos(), "data sources should return an error when a resource cannot be found instead of calling SetId with an empty string")
	})

	return nil, nil
}
