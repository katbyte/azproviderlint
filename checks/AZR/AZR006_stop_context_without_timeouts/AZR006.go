// Package AZR006 defines an analyzer that reports resources assigning ctx directly from the
// provider meta object instead of using a timeouts-wrapped StopContext.
package AZR006

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `ctx := meta.(*clients.Client).StopContext` style assignments. To
// support Custom Timeouts the StopContext must be wrapped, e.g.
// `ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)` with a
// corresponding `defer cancel()` (ForCreate/ForCreateUpdate/ForRead/ForUpdate/ForDelete).
var Analyzer = &analysis.Analyzer{
	Name:     "AZR006",
	Doc:      "check for ctx being assigned directly from the provider meta object instead of a timeouts-wrapped StopContext",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR006_stop_context_without_timeouts/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != token.DEFINE {
			return
		}

		if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return
		}

		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhs.Name != "ctx" {
			return
		}

		root := rootIdent(assign.Rhs[0])
		if root == nil || root.Name != "meta" {
			return
		}

		pass.Reportf(assign.Pos(), "use a timeouts-wrapped StopContext (timeouts.ForCreate/ForCreateUpdate/ForRead/ForUpdate/ForDelete) so Custom Timeouts are supported, instead of assigning ctx from meta directly")
	})

	return nil, nil
}

// rootIdent walks selector/type-assertion/call/index chains down to the leftmost identifier.
func rootIdent(expr ast.Expr) *ast.Ident {
	for {
		switch v := expr.(type) {
		case *ast.Ident:
			return v
		case *ast.SelectorExpr:
			expr = v.X
		case *ast.TypeAssertExpr:
			expr = v.X
		case *ast.CallExpr:
			expr = v.Fun
		case *ast.ParenExpr:
			expr = v.X
		case *ast.IndexExpr:
			expr = v.X
		default:
			return nil
		}
	}
}
