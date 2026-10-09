// Package AZR004 defines an analyzer that reports Resource IDs being compared with the
// == or != operators instead of resourceids.Match.
package AZR004

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `a.ID() == b.ID()` style comparisons. Azure Resource IDs have
// case-insensitive segments, so string equality is unreliable - use `resourceids.Match(a, b)`
// from github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids instead.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR004",
	Doc:      "check for Resource IDs being compared with == or != instead of resourceids.Match",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR004_resource_id_equality_comparison/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.BinaryExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		cmp, ok := n.(*ast.BinaryExpr)
		if !ok || (cmp.Op != token.EQL && cmp.Op != token.NEQ) {
			return
		}

		if !isIDCall(cmp.X) && !isIDCall(cmp.Y) {
			return
		}

		pass.Reportf(cmp.Pos(), "Resource IDs should not be compared with == or !=, use resourceids.Match instead")
	})

	return nil, nil
}

// isIDCall reports whether the expression is a call to a niladic ID() method.
func isIDCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "ID"
}
