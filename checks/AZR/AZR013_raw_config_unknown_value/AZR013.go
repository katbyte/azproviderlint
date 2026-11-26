// Package AZR013 defines an analyzer that reports a raw config value read while planning with no check that it is known yet.
package AZR013

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks the code that reads the raw config of a plan. A value that comes from a resource not created yet is not known while planning,
// and asking such a value for its length, its elements or its text panics, so each one has to be asked whether it is known first.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR013",
	Doc:      "check that a raw config value read while planning is checked with IsKnown first",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR013_raw_config_unknown_value/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// reads are the methods of a cty.Value that panic when the value is not known.
var reads = []string{"LengthInt", "AsValueSlice", "AsValueMap", "AsValueSet", "ElementIterator", "AsString", "AsBigFloat", "True", "False"}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, isDecl := n.(*ast.FuncDecl)
		if !isDecl || decl.Body == nil {
			return
		}

		// what is read, where a value is asked whether it is known, and whether any of it is the config of a plan
		var read []*ast.SelectorExpr
		known, planning, wholly := map[string]token.Pos{}, false, false
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall {
				return true
			}

			sel, isMethod := call.Fun.(*ast.SelectorExpr)
			if !isMethod || pass.TypesInfo.TypeOf(sel.X) == nil {
				return true
			}

			on := pass.TypesInfo.TypeOf(sel.X)
			if ptr, isPointer := types.Unalias(on).(*types.Pointer); isPointer {
				on = ptr.Elem()
			}

			named, isNamed := types.Unalias(on).(*types.Named)
			if !isNamed || named.Obj().Pkg() == nil {
				return true
			}

			switch {
			case named.Obj().Name() == "ResourceDiff" && (sel.Sel.Name == "GetRawConfig" || sel.Sel.Name == "GetRawPlan"):
				planning = true

			case named.Obj().Name() != "Value" || named.Obj().Pkg().Name() != "cty":

			case sel.Sel.Name == "IsWhollyKnown":
				wholly = true

			case sel.Sel.Name == "IsKnown":
				if _, seen := known[types.ExprString(sel.X)]; !seen {
					known[types.ExprString(sel.X)] = call.Pos()
				}

			case slices.Contains(reads, sel.Sel.Name):
				read = append(read, sel)
			}

			return true
		})

		// a value that is wholly known has nothing in it that is not
		if !planning || wholly {
			return
		}

		for _, sel := range read {
			value := types.ExprString(sel.X)
			if asked, isAsked := known[value]; !isAsked || asked > sel.Pos() {
				pass.Reportf(sel.Pos(), "%s.%s() panics while planning when the value is not known yet - check %s.IsKnown() first", value, sel.Sel.Name, value)
			}
		}
	})

	return nil, nil
}
