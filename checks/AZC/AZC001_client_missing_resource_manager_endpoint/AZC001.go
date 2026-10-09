// Package AZC001 defines an analyzer that reports Azure SDK clients being created without
// the resource manager endpoint explicitly specified.
package AZC001

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `NewFoosClient(o.SubscriptionId)` style constructions. Azure SDK
// (track1 & kermit) clients should be created with `NewFoosClientWithBaseURI(...)` so the
// resource manager endpoint is explicitly specified (required for sovereign clouds).
var Analyzer = &analysis.Analyzer{
	Name:     "AZC001",
	Doc:      "check for Azure SDK clients created without the resource manager endpoint explicitly specified via NewFoosClientWithBaseURI",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZC/AZC001_client_missing_resource_manager_endpoint/README.md",
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
		if !ok || len(call.Args) != 1 {
			return
		}

		if !strings.HasSuffix(calleeName(call.Fun), "Client") {
			return
		}

		arg, ok := call.Args[0].(*ast.SelectorExpr)
		if !ok || arg.Sel.Name != "SubscriptionId" {
			return
		}

		recv, ok := arg.X.(*ast.Ident)
		if !ok || recv.Name != "o" {
			return
		}

		pass.Reportf(call.Pos(), "Azure SDK clients should be created with NewFoosClientWithBaseURI and the resource manager endpoint explicitly specified")
	})

	return nil, nil
}

// calleeName returns the called function's name, unwrapping package selectors.
func calleeName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	}
	return ""
}
