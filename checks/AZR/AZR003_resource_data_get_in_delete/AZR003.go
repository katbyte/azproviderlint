// Package AZR003 defines an analyzer that reports ResourceData.Get being used inside a
// resource's Delete function, where it does not work as expected.
package AZR003

import (
	"go/ast"

	"github.com/katbyte/azproviderlint/lib/lifecycle"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks Delete functions for schema data reads. During deletion the state may be
// partial, so `d.Get(...)` (untyped resources, found via the `Delete:` registration) and
// `metadata.ResourceData.Get(...)` (typed resources, inside `Delete() sdk.ResourceFunc`)
// do not behave as expected and should not be used.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR003",
	Doc:      "check for ResourceData.Get being used inside a resource's Delete function where it does not work as expected",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR003_resource_data_get_in_delete/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	for fn, step := range lifecycle.Funcs(insp) {
		if step != lifecycle.Delete {
			continue
		}
		if fn.Recv == nil {
			// untyped resources: the function registered via `Delete: resourceFooDelete,`
			checkUntypedDelete(pass, fn)
		} else {
			// typed resources: the `Delete() sdk.ResourceFunc` method
			checkTypedDelete(pass, fn)
		}
	}

	return nil, nil
}

// checkUntypedDelete reports `d.Get(...)` calls, where `d` is the function's first parameter.
func checkUntypedDelete(pass *analysis.Pass, fn *ast.FuncDecl) {
	dataParam := firstParamName(fn)
	if dataParam == "" {
		return
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Get" {
			return true
		}

		recv, ok := sel.X.(*ast.Ident)
		if !ok || recv.Name != dataParam {
			return true
		}

		pass.Reportf(call.Pos(),
			"%s.Get should not be used within a Delete function as it does not work as expected during deletion", dataParam)
		return true
	})
}

// checkTypedDelete reports `metadata.ResourceData.Get(...)` calls.
func checkTypedDelete(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Get" {
			return true
		}

		inner, ok := sel.X.(*ast.SelectorExpr)
		if !ok || inner.Sel.Name != "ResourceData" {
			return true
		}

		pass.Reportf(call.Pos(),
			"ResourceData.Get should not be used within a Delete function as it does not work as expected during deletion")
		return true
	})
}

func firstParamName(fn *ast.FuncDecl) string {
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 || len(fn.Type.Params.List[0].Names) == 0 {
		return ""
	}
	return fn.Type.Params.List[0].Names[0].Name
}
