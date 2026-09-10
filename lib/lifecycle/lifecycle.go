// Package lifecycle finds the functions that implement a resource's Create, Read, Update, and
// Delete steps, in both provider styles.
package lifecycle

import (
	"go/ast"

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
// names returning `sdk.ResourceFunc`. A function registered under two keys (a combined
// CreateUpdate) is reported under the first seen.
func Funcs(insp *inspector.Inspector) map[*ast.FuncDecl]string {
	registered := map[string]string{} // function name -> step
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
		var name string
		switch v := kv.Value.(type) {
		case *ast.Ident:
			name = v.Name
		case *ast.SelectorExpr:
			name = v.Sel.Name
		default:
			return
		}
		if _, seen := registered[name]; !seen {
			registered[name] = step
		}
	})

	funcs := map[*ast.FuncDecl]string{}
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return
		}
		if fn.Recv == nil {
			if step, ok := registered[fn.Name.Name]; ok {
				funcs[fn] = step
			}
			return
		}
		if step, ok := steps[fn.Name.Name]; ok && returnsResourceFunc(fn) {
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
