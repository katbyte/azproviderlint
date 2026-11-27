// Package AZR012 defines an analyzer that reports a create that refuses a resource that already exists without asking whether the provider is
// set to overwrite it.
package AZR012

import (
	"go/ast"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that the error telling the user to import a resource that already exists is only returned inside a check of the feature that
// lets create overwrite it, so a user who has switched that feature on is not refused anyway.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR012",
	Doc:      "check that the already-exists error on create sits inside a check of the overwrite feature flag",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR012_import_check_ignores_overwrite_flag/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// flag is the provider feature that lets create overwrite a resource that is already there.
const flag = "SkipImportCheckOnCreateAndAllowOverwritingExistingResources"

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	insp.WithStack([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		call, isCall := n.(*ast.CallExpr)
		if !push || !isCall {
			return true
		}

		// the untyped and the typed way of saying a resource already exists and has to be imported
		fn := astx.CalledFunc(pass, call)
		if fn == nil || (fn.Name() != "ImportAsExistsError" && fn.Name() != "ResourceRequiresImport") {
			return true
		}

		for _, outer := range stack {
			switch outer := outer.(type) {
			case *ast.FuncDecl:
				// the function that builds the error for the other is not a create
				if outer.Name.Name == "ImportAsExistsError" || outer.Name.Name == "ResourceRequiresImport" {
					return true
				}

			case *ast.IfStmt:
				guarded := false
				ast.Inspect(outer.Cond, func(n ast.Node) bool {
					id, isIdent := n.(*ast.Ident)
					guarded = guarded || (isIdent && id.Name == flag)

					return !guarded
				})
				if guarded {
					return true
				}
			}
		}

		pass.Reportf(call.Pos(), "this already-exists error is returned even when the provider is set to overwrite - put the check inside `if !<client>.Features.%s`", flag)

		return true
	})

	return nil, nil
}
