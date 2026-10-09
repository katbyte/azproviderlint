// Package AZR002 defines an analyzer that reports resources registering a combined
// CreateUpdate method instead of separate Create and Update methods.
package AZR002

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for `Create: resourceFooCreateUpdate,` registrations. New resources
// should define separate Create and Update methods; combined CreateUpdate methods are
// gradually being split across the provider.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR002",
	Doc:      "check for resources registering a combined CreateUpdate method instead of separate Create and Update methods",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR002_combined_create_update_method/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.KeyValueExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Create" {
			return
		}

		if !strings.HasSuffix(valueName(kv.Value), "CreateUpdate") {
			return
		}

		pass.Reportf(kv.Pos(), "new resources should use separate Create and Update methods instead of a combined CreateUpdate method")
	})

	return nil, nil
}

// valueName returns the identifier name a Create: entry refers to, unwrapping selectors.
func valueName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	}
	return ""
}
