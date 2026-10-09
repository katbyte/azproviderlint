// Package AZR001 defines an analyzer that reports SetId being called with a dereferenced
// pointer (typically the raw Azure API resource ID) instead of a generated Resource ID
// Formatter/Parser's id.ID().
package AZR001

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `d.SetId(*read.ID)` style calls. The Azure API returns Resource IDs
// inconsistently, so Terraform manages its own Resource IDs - new resources should build
// the ID with a generated Resource ID Formatter and call `d.SetId(id.ID())`.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR001",
	Doc:      "check for SetId being called with a dereferenced pointer instead of a Resource ID Formatter/Parser's id.ID()",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR001_set_id_dereferenced_pointer/README.md",
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
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "SetId" || len(call.Args) != 1 {
			return
		}

		if _, ok := call.Args[0].(*ast.StarExpr); !ok {
			return
		}

		pass.Reportf(call.Pos(), "SetId should not be passed a dereferenced pointer, use a generated Resource ID Formatter/Parser and id.ID()")
	})

	return nil, nil
}
