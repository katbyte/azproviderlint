// Package AZT011 defines an analyzer that reports an acceptance test callback that polls with the context it was handed, which has no deadline.
package AZT011

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks the callbacks a test hands to CheckWithClient and its siblings. They are given the provider's own context, which never ends,
// and a poller refuses a context with no deadline, so the test fails before it has waited at all.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT011",
	Doc:      "check that a test callback gives its context a deadline before polling with it",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT011_test_callback_poll_without_deadline/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// checks are the test helpers that call back with the provider's context as it is.
var checks = []string{"CheckWithClient", "CheckWithClientForResource", "CheckWithClientWithoutResource"}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// a callback can be a function of the package, named where the check is made
	decls := map[types.Object]*ast.FuncDecl{}
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		if decl, isDecl := n.(*ast.FuncDecl); isDecl && decl.Body != nil {
			decls[pass.TypesInfo.Defs[decl.Name]] = decl
		}
	})

	seen := map[*ast.BlockStmt]bool{}
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, isCall := n.(*ast.CallExpr)
		if !isCall || len(call.Args) == 0 {
			return
		}

		if fn := astx.CalledFunc(pass, call); fn == nil || fn.Signature().Recv() == nil || !slices.Contains(checks, fn.Name()) {
			return
		}

		var params *ast.FieldList
		var body *ast.BlockStmt
		switch callback := ast.Unparen(call.Args[0]).(type) {
		case *ast.FuncLit:
			params, body = callback.Type.Params, callback.Body

		case *ast.Ident:
			if decl := decls[pass.TypesInfo.Uses[callback]]; decl != nil {
				params, body = decl.Type.Params, decl.Body
			}

		case *ast.SelectorExpr:
			if decl := decls[pass.TypesInfo.Uses[callback.Sel]]; decl != nil {
				params, body = decl.Type.Params, decl.Body
			}
		}

		// the context is the first thing a callback is handed
		if body == nil || seen[body] || len(params.List) == 0 || len(params.List[0].Names) == 0 {
			return
		}

		seen[body] = true
		ctx := pass.TypesInfo.Defs[params.List[0].Names[0]]
		if ctx == nil || ctx.Type().String() != "context.Context" {
			return
		}

		// from where the context is swapped for one with a deadline, polling with it is fine
		bounded := token.NoPos
		ast.Inspect(body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				if len(n.Rhs) != 1 || len(n.Lhs) == 0 || !astx.IsUseOf(pass, n.Lhs[0], ctx) {
					return true
				}

				wrap, isWrap := n.Rhs[0].(*ast.CallExpr)
				if !isWrap {
					return true
				}

				if fn := astx.CalledFunc(pass, wrap); fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == "context" && (fn.Name() == "WithTimeout" || fn.Name() == "WithDeadline") && !bounded.IsValid() {
					bounded = n.End()
				}

			case *ast.CallExpr:
				fn := astx.CalledFunc(pass, n)
				if fn == nil || (!strings.HasSuffix(fn.Name(), "ThenPoll") && fn.Name() != "PollUntilDone") || (bounded.IsValid() && n.Pos() > bounded) {
					return true
				}

				for _, arg := range n.Args {
					if astx.IsUseOf(pass, arg, ctx) {
						pass.Reportf(arg.Pos(), "this callback's context has no deadline, and %s refuses to poll without one - wrap it with context.WithTimeout first", fn.Name())
					}
				}
			}

			return true
		})
	})

	return nil, nil
}
