// Package AZR015 defines an analyzer that reports a create that hands the API a callback setting the ID of a resource but not its identity.
package AZR015

import (
	"go/ast"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks the callback a create gives a long-running API call, which saves the resource before the call has finished. A resource with an
// identity has to save that too, or Terraform fails the create for leaving it out.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR015",
	Doc:      "check that the create callback of a resource with an identity sets the identity as well as the ID",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR015_create_callback_missing_identity/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// the creates of the resources that have an identity: named by a resource that sets Identity, or a method of a type with an Identity method
	creates := map[types.Object]bool{}
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || !tf.IsSchemaHelperType(pass, lit, "Resource") {
			return
		}

		fields := map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			if kv, isPair := elt.(*ast.KeyValueExpr); isPair {
				if key, isIdent := kv.Key.(*ast.Ident); isIdent {
					fields[key.Name] = kv.Value
				}
			}
		}

		if fields["Identity"] == nil || astx.IsNilValue(pass, fields["Identity"]) {
			return
		}

		for _, name := range []string{"Create", "CreateContext", "CreateWithoutTimeout"} {
			if create, isIdent := fields[name].(*ast.Ident); isIdent {
				creates[pass.TypesInfo.Uses[create]] = true
			}
		}
	})

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, isDecl := n.(*ast.FuncDecl)
		if !isDecl || decl.Body == nil {
			return
		}

		fn, isFunc := pass.TypesInfo.Defs[decl.Name].(*types.Func)
		if !isFunc {
			return
		}

		if recv := fn.Signature().Recv(); recv != nil && decl.Name.Name == "Create" {
			identity, _, _ := types.LookupFieldOrMethod(recv.Type(), true, pass.Pkg, "Identity")
			_, creates[fn] = identity.(*types.Func)
		}

		if !creates[fn] {
			return
		}

		ast.Inspect(decl.Body, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall {
				return true
			}

			called := astx.CalledFunc(pass, call)
			if called == nil || called.Name() != "SetIDCallback" {
				return true
			}

			// the name is the last part of sdk.SetIDCallback and of metadata.SetIDCallback alike
			name := call.Fun
			if sel, isSel := name.(*ast.SelectorExpr); isSel {
				name = sel.Sel
			}

			pass.Report(analysis.Diagnostic{Pos: name.Pos(), End: name.End(), Message: "this resource has an identity, and SetIDCallback sets only the ID - use SetIDAndIdentityCallback", SuggestedFixes: []analysis.SuggestedFix{{Message: "set the identity too", TextEdits: []analysis.TextEdit{{Pos: name.Pos(), End: name.End(), NewText: []byte("SetIDAndIdentityCallback")}}}}})

			return true
		})
	})

	return nil, nil
}
