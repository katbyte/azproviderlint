// Package description finds the descriptions written out in a provider's code, so every check on them looks in the same places.
package description

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// Found is one description written as text in the code: on a schema property, a resource, a framework attribute or block, or returned by a typed
// resource's Description() method.
type Found struct {
	// Value is the expression holding the text.
	Value ast.Expr

	// Text is what it says.
	Text string
}

// All returns every description in the package, outside its test files, that is text the compiler can read: a string, or constants joined together.
func All(pass *analysis.Pass, insp *inspector.Inspector) []Found {
	var found []Found
	add := func(value ast.Expr) {
		if tv, ok := pass.TypesInfo.Types[value]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
			found = append(found, Found{Value: value, Text: constant.StringVal(tv.Value)})
		}
	}

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil), (*ast.FuncDecl)(nil)}, func(n ast.Node) {
		if strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return
		}

		switch n := n.(type) {
		case *ast.CompositeLit:
			t := pass.TypesInfo.TypeOf(n)
			if t == nil {
				return
			}
			if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
				t = ptr.Elem()
			}

			// the plugin sdk and the framework both keep their schema types in a package named schema; an Azure API model with a Description
			// field is not one
			named, ok := types.Unalias(t).(*types.Named)
			if !ok || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" || !strings.Contains(named.Obj().Pkg().Path(), "/terraform-plugin-") {
				return
			}

			for _, elt := range n.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				if key, ok := kv.Key.(*ast.Ident); ok && (key.Name == "Description" || key.Name == "MarkdownDescription") {
					add(kv.Value)
				}
			}

		case *ast.FuncDecl:
			if n.Recv == nil || n.Name.Name != "Description" || n.Body == nil || len(n.Body.List) != 1 {
				return
			}

			fn, ok := pass.TypesInfo.Defs[n.Name].(*types.Func)
			if !ok {
				return
			}

			sig := fn.Signature()
			if sig.Params().Len() != 0 || sig.Results().Len() != 1 || !types.Identical(sig.Results().At(0).Type(), types.Typ[types.String]) {
				return
			}

			// every type the sdk registers says what it is called, which tells its Description() from any other
			if obj, _, _ := types.LookupFieldOrMethod(sig.Recv().Type(), true, pass.Pkg, "ResourceType"); obj == nil {
				return
			}

			if ret, ok := n.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
				add(ret.Results[0])
			}
		}
	})

	return found
}
