// Package AZT012 defines an analyzer that reports an acceptance test step that applies a config but never checks the resource exists, or is not
// followed by an import step.
package AZT012

import (
	"go/ast"
	"go/constant"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks every step of a resource's acceptance tests that applies a config. Applying a config proves only that Terraform did not fail:
// the exists check proves the resource was made, and the import step after it proves what was saved matches what the API gives back.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT012",
	Doc:      "check that each step of a resource's acceptance test checks the resource exists and is followed by an import step",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT012_test_step_missing_exists_or_import/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// called gives the name of the function a call runs, or nothing when it is not a call
	called := func(e ast.Expr) string {
		call, isCall := e.(*ast.CallExpr)
		if !isCall {
			return ""
		}

		if fn := astx.CalledFunc(pass, call); fn != nil {
			return fn.Name()
		}

		return ""
	}

	// fields gives the fields a step written out in place sets, by name
	fields := func(e ast.Expr) map[string]ast.Expr {
		lit, isLit := e.(*ast.CompositeLit)
		if !isLit {
			return nil
		}

		set := map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			if kv, isPair := elt.(*ast.KeyValueExpr); isPair {
				if key, isField := kv.Key.(*ast.Ident); isField {
					set[key.Name] = kv.Value
				}
			}
		}

		return set
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, isDecl := n.(*ast.FuncDecl)
		if !isDecl || decl.Body == nil || decl.Recv != nil || !strings.HasPrefix(decl.Name.Name, "TestAcc") || strings.Contains(decl.Name.Name, "DataSource") {
			return
		}

		// the steps of a resource test, when they are written out where the test is run; a data source has nothing to exist or import
		var steps *ast.CompositeLit
		dataSource := false
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall || len(call.Args) == 0 {
				return true
			}

			fn := astx.CalledFunc(pass, call)
			if fn == nil {
				return true
			}

			if of, isText := pass.TypesInfo.Types[call.Args[min(1, len(call.Args)-1)]]; fn.Name() == "BuildTestData" && isText && of.Value != nil && of.Value.Kind() == constant.String {
				dataSource = dataSource || strings.HasPrefix(constant.StringVal(of.Value), "data.")
			}

			if fn.Signature().Recv() != nil && strings.HasPrefix(fn.Name(), "Resource") && strings.Contains(fn.Name(), "Test") && steps == nil {
				steps, _ = call.Args[len(call.Args)-1].(*ast.CompositeLit)
			}

			return true
		})

		if steps == nil || dataSource {
			return
		}

		for i, step := range steps.Elts {
			set := fields(step)

			// a step that applies a config: one written out with a config, or ApplyStep, which also checks the resource exists
			applies := called(step) == "ApplyStep"
			if set != nil {
				applies = set["Config"] != nil && set["ExpectError"] == nil && !astx.IsTrueConstant(pass, set["PlanOnly"]) && !astx.IsTrueConstant(pass, set["Destroy"]) && !astx.IsTrueConstant(pass, set["ImportState"])
			}
			if !applies {
				continue
			}

			// a step that checks the resource is gone has nothing to exist or import
			exists, gone := called(step) == "ApplyStep", false
			ast.Inspect(step, func(n ast.Node) bool {
				if call, isCall := n.(*ast.CallExpr); isCall {
					gone, exists = gone || strings.Contains(called(call), "NotExist"), exists || (strings.Contains(called(call), "Exists") && !strings.Contains(called(call), "NotExist"))
				}

				return true
			})
			if gone {
				continue
			}

			// what follows imports the resource, or tries to create it again to show the provider refuses to
			var next ast.Expr
			if i+1 < len(steps.Elts) {
				next = steps.Elts[i+1]
			}
			after := fields(next)
			imports := next != nil && (strings.HasPrefix(called(next), "Import") || called(next) == "RequiresImportErrorStep" || astx.IsTrueConstant(pass, after["ImportState"]) || called(after["ExpectError"]) == "RequiresImportError")

			if !exists {
				pass.Reportf(step.Pos(), "step %d of %s never checks the resource exists - add check.That(data.ResourceName).ExistsInAzure(r) to its Check", i+1, decl.Name.Name)
			}
			if !imports {
				pass.Reportf(step.Pos(), "step %d of %s is not followed by an import step - add data.ImportStep() after it", i+1, decl.Name.Name)
			}
		}
	})

	return nil, nil
}
