// Package acctest finds the functions of a package that run an acceptance test, for the
// checks that compare that against how the function is named.
package acctest

import (
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/astx"
)

// Analyzer computes Funcs for a package. A function runs an acceptance test when it hands
// `t` to the Terraform test harness or the provider's own internal/acceptance package, looks
// at TF_ACC, or calls or passes around another function of the package that does. The
// harness's UnitTest entry point is the exception: it runs without TF_ACC by design. Packages
// without test files have nothing to find and yield an empty result.
var Analyzer = &analysis.Analyzer{
	Name:       "acctest",
	Doc:        "find the functions that run an acceptance test",
	Run:        run,
	ResultType: reflect.TypeFor[Funcs](),
}

// Funcs maps every function in the package that runs an acceptance test to its declaration.
type Funcs map[*types.Func]*ast.FuncDecl

// The Terraform test harnesses; a provider's own internal/acceptance package counts too.
var harnessPackages = map[string]bool{
	"github.com/hashicorp/terraform-plugin-testing/helper/resource": true,
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource":  true,
}

func run(pass *analysis.Pass) (any, error) {
	hasTests := slices.ContainsFunc(pass.Files, func(f *ast.File) bool {
		return strings.HasSuffix(pass.Fset.Position(f.Pos()).Filename, "_test.go")
	})
	if !hasTests {
		return Funcs{}, nil
	}

	decls := map[*types.Func]*ast.FuncDecl{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			if fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func); ok {
				decls[fn] = fd
			}
		}
	}

	// runs marks the functions that are acceptance tests on their own: they hand t to a
	// harness or look at TF_ACC. mentions lists the same-package functions each one calls or
	// passes around as a value (`t.Run("basic", testAccThing_basic)`).
	runs := map[*types.Func]bool{}
	mentions := map[*types.Func][]*types.Func{}
	for fn, fd := range decls {
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				callee := astx.CalledFunc(pass, n)
				if callee == nil || callee.Pkg() == nil {
					break
				}
				if path := callee.Pkg().Path(); (!harnessPackages[path] && !strings.HasSuffix(path, "/internal/acceptance")) || callee.Name() == "UnitTest" {
					break
				}
				for _, arg := range n.Args {
					ptr, ok := pass.TypesInfo.TypeOf(arg).(*types.Pointer)
					if !ok {
						continue
					}
					if named, ok := ptr.Elem().(*types.Named); ok && named.Obj().Name() == "T" && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "testing" {
						runs[fn] = true
					}
				}
			case *ast.Ident:
				if used, ok := pass.TypesInfo.Uses[n].(*types.Func); ok && used.Pkg() == pass.Pkg {
					mentions[fn] = append(mentions[fn], used.Origin())
				}
			}
			// by value, so resource.EnvTfAcc and local constants count as well as the literal
			if e, ok := n.(ast.Expr); ok {
				if v := pass.TypesInfo.Types[e].Value; v != nil && v.Kind() == constant.String && constant.StringVal(v) == "TF_ACC" {
					runs[fn] = true
				}
			}
			return true
		})
	}

	// a function that mentions an acceptance test is one too
	for changed := true; changed; {
		changed = false
		for fn, used := range mentions {
			if !runs[fn] && slices.ContainsFunc(used, func(u *types.Func) bool { return runs[u] }) {
				runs[fn] = true
				changed = true
			}
		}
	}

	found := Funcs{}
	for fn := range runs {
		found[fn] = decls[fn]
	}

	return found, nil
}
