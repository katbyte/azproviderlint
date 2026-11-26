// Package AZT012 defines an analyzer that reports a basic acceptance test that never checks the resource exists, or never imports it.
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

// Analyzer checks the basic test of a resource, the one every other test builds on. Applying a config proves only that Terraform did not fail:
// the exists check proves the resource was made, and the import step proves what was saved matches what the API gives back.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT012",
	Doc:      "check that a resource's basic acceptance test checks the resource exists and imports it",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT012_basic_test_missing_exists_or_import/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, isDecl := n.(*ast.FuncDecl)
		if !isDecl || decl.Body == nil || decl.Recv != nil || !strings.HasPrefix(decl.Name.Name, "TestAcc") || !strings.HasSuffix(decl.Name.Name, "_basic") || strings.Contains(decl.Name.Name, "DataSource") {
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

		exists, imports := false, false
		ast.Inspect(steps, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				if fn := astx.CalledFunc(pass, n); fn != nil {
					// ApplyStep is a step that applies a config and checks the resource exists
					exists, imports = exists || fn.Name() == "ExistsInAzure" || fn.Name() == "ApplyStep", imports || strings.HasPrefix(fn.Name(), "Import")
				}

			case *ast.KeyValueExpr:
				if key, isField := n.Key.(*ast.Ident); isField && key.Name == "ImportState" {
					imports = imports || astx.IsTrueConstant(pass, n.Value)
				}
			}

			return true
		})

		if !exists {
			pass.Reportf(decl.Name.Pos(), "%s never checks the resource exists - add check.That(data.ResourceName).ExistsInAzure(r) to a step", decl.Name.Name)
		}
		if !imports {
			pass.Reportf(decl.Name.Pos(), "%s never imports the resource - add data.ImportStep() after the config is applied", decl.Name.Name)
		}
	})

	return nil, nil
}
