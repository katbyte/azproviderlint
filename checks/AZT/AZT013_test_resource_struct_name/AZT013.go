// Package AZT013 defines an analyzer that reports an acceptance test's resource struct that is not named after the resource it tests.
package AZT013

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks the name of the struct a resource's acceptance tests hang off. Generated tests, such as the identity ones, work the name out
// from the resource's, so a struct named any other way is one they cannot find.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT013",
	Doc:      "check that an acceptance test's resource struct is named after the resource",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT013_test_resource_struct_name/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// each struct handed to a test runner, the resources it is tested as, and where it was first handed over
	var structs []types.Object
	resources, first := map[types.Object][]string{}, map[types.Object]token.Pos{}
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, isDecl := n.(*ast.FuncDecl)
		if !isDecl || decl.Body == nil {
			return
		}

		// a test says which resource it is for where it builds its data, and hands its struct to whatever runs the steps
		resource := ""
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall || len(call.Args) < 2 {
				return true
			}

			fn := astx.CalledFunc(pass, call)
			if fn == nil {
				return true
			}

			if name, isText := pass.TypesInfo.Types[call.Args[1]]; fn.Name() == "BuildTestData" && resource == "" && isText && name.Value != nil && name.Value.Kind() == constant.String {
				resource = constant.StringVal(name.Value)
			}

			if fn.Signature().Recv() == nil || !strings.HasPrefix(fn.Name(), "Resource") || !strings.Contains(fn.Name(), "Test") || resource == "" || strings.HasPrefix(resource, "data.") || pass.TypesInfo.TypeOf(call.Args[1]) == nil {
				return true
			}

			t := pass.TypesInfo.TypeOf(call.Args[1])
			if ptr, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
				t = ptr.Elem()
			}

			named, isNamed := types.Unalias(t).(*types.Named)
			if !isNamed {
				return true
			}

			if _, seen := first[named.Obj()]; !seen {
				structs, first[named.Obj()] = append(structs, named.Obj()), call.Args[1].Pos()
			}
			resources[named.Obj()] = append(resources[named.Obj()], resource)

			return true
		})
	})

	for _, tested := range structs {
		// azurerm_storage_account is tested by StorageAccountResource: the provider's name dropped, each word given a capital
		var wanted []string
		for _, resource := range resources[tested] {
			_, rest, _ := strings.Cut(resource, "_")
			var name strings.Builder
			for word := range strings.SplitSeq(rest, "_") {
				if word != "" {
					name.WriteString(strings.ToUpper(word[:1]) + word[1:])
				}
			}

			wanted = append(wanted, name.String()+"Resource")
		}

		// a struct shared by several resources can only be named after one of them
		if slices.Contains(wanted, tested.Name()) {
			continue
		}

		at := first[tested]
		if tested.Pkg() == pass.Pkg {
			at = tested.Pos()
		}

		pass.Reportf(at, "test resource %s is not named after %s, which it tests - name it %s", tested.Name(), resources[tested][0], wanted[0])
	}

	return nil, nil
}
