// Package lifecycle finds the functions that implement a resource's Create, Read, Update, and
// Delete steps, in both provider styles.
package lifecycle

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// The lifecycle steps, as Funcs names them.
const (
	Create = "Create"
	Read   = "Read"
	Update = "Update"
	Delete = "Delete"
)

// Funcs returns every function in the package that implements a lifecycle step, with the
// step's name: untyped resources' functions registered under a `Create:`, `Read:`, `Update:`,
// or `Delete:` key (the `*Context` spellings too), and typed resources' methods of those
// names returning `sdk.ResourceFunc`. Registered values are resolved through type
// information, so a handler from another package or a function literal has no declaration
// here and is skipped. A function registered under two keys (a combined CreateUpdate) is
// reported under the first seen.
func Funcs(pass *analysis.Pass, insp *inspector.Inspector) map[*ast.FuncDecl]string {
	registered := map[types.Object]string{} // function -> step
	insp.Preorder([]ast.Node{(*ast.KeyValueExpr)(nil)}, func(n ast.Node) {
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			return
		}
		step, ok := steps[key.Name]
		if !ok {
			return
		}
		var id *ast.Ident
		switch v := kv.Value.(type) {
		case *ast.Ident:
			id = v
		case *ast.SelectorExpr:
			id = v.Sel
		default:
			return
		}
		fn, ok := pass.TypesInfo.Uses[id].(*types.Func)
		if !ok {
			return
		}
		if _, seen := registered[fn]; !seen {
			registered[fn] = step
		}
	})

	funcs := map[*ast.FuncDecl]string{}
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return
		}
		if step, ok := registered[pass.TypesInfo.Defs[fn.Name]]; ok {
			funcs[fn] = step
			return
		}
		if step, ok := steps[fn.Name.Name]; ok && fn.Recv != nil && returnsResourceFunc(fn) {
			funcs[fn] = step
		}
	})
	return funcs
}

// steps maps a registration key or typed method name to its lifecycle step.
var steps = map[string]string{
	Create: Create, Create + "Context": Create,
	Read: Read, Read + "Context": Read,
	Update: Update, Update + "Context": Update,
	Delete: Delete, Delete + "Context": Delete,
}

// returnsResourceFunc reports whether the function's single result type is (sdk.)ResourceFunc.
func returnsResourceFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	switch t := fn.Type.Results.List[0].Type.(type) {
	case *ast.Ident:
		return t.Name == "ResourceFunc"
	case *ast.SelectorExpr:
		return t.Sel.Name == "ResourceFunc"
	}
	return false
}
