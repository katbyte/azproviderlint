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

	// check looks at one function. planning says it is known to run while planning, which a function that reads the config of a plan shows by itself.
	at := map[token.Pos]bool{}
	check := func(body *ast.BlockStmt, planning bool) {
		// config reports whether an expression is the whole config, and item whether it is one value out of a list or set
		config := func(e ast.Expr) bool {
			call, isCall := ast.Unparen(e).(*ast.CallExpr)
			if !isCall {
				return false
			}

			sel, isMethod := call.Fun.(*ast.SelectorExpr)

			return isMethod && (sel.Sel.Name == "GetRawConfig" || sel.Sel.Name == "GetRawPlan")
		}
		item := func(e ast.Expr) bool {
			if index, isIndex := ast.Unparen(e).(*ast.IndexExpr); isIndex {
				e = index.X
			}

			call, isCall := ast.Unparen(e).(*ast.CallExpr)
			if !isCall {
				return false
			}

			sel, isMethod := call.Fun.(*ast.SelectorExpr)

			return isMethod && (sel.Sel.Name == "AsValueSlice" || sel.Sel.Name == "Values")
		}

		// the names given to the whole config, and to one value out of a list or set
		whole, items := map[types.Object]bool{}, map[types.Object]bool{}
		ast.Inspect(body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				if name, isName := n.Lhs[0].(*ast.Ident); isName && len(n.Lhs) == 1 && len(n.Rhs) == 1 {
					_, isIndex := ast.Unparen(n.Rhs[0]).(*ast.IndexExpr)
					whole[pass.TypesInfo.ObjectOf(name)], items[pass.TypesInfo.ObjectOf(name)] = config(n.Rhs[0]), isIndex && item(n.Rhs[0])
				}

			case *ast.RangeStmt:
				if name, isName := n.Value.(*ast.Ident); isName && item(n.X) {
					items[pass.TypesInfo.ObjectOf(name)] = true
				}
			}

			return true
		})

		// what is read, where a value is asked whether it is known, and whether any of it is the config of a plan
		var read []*ast.SelectorExpr
		known, wholly := map[string]token.Pos{}, false
		ast.Inspect(body, func(n ast.Node) bool {
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
				// the config as a whole is always known, and so is a block in a list that is: only what a block holds can be not known yet
				name, isName := ast.Unparen(sel.X).(*ast.Ident)
				if config(sel.X) || (isName && whole[pass.TypesInfo.ObjectOf(name)]) || (sel.Sel.Name == "AsValueMap" && (item(sel.X) || (isName && items[pass.TypesInfo.ObjectOf(name)]))) {
					return true
				}

				read = append(read, sel)
			}

			return true
		})

		// a value that is wholly known has nothing in it that is not
		if !planning || wholly {
			return
		}

		// a value read twice with no check is reported where it is first read
		reported := map[string]bool{}
		for _, sel := range read {
			value := types.ExprString(sel.X)
			if asked, isAsked := known[value]; (!isAsked || asked > sel.Pos()) && !reported[value] && !at[sel.Pos()] {
				reported[value], at[sel.Pos()] = true, true
				pass.Reportf(sel.Pos(), "%s.%s() panics while planning when the value is not known yet - check %s.IsKnown() first", value, sel.Sel.Name, value)
			}
		}
	}

	decls := map[types.Object]*ast.FuncDecl{}
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		if decl, isDecl := n.(*ast.FuncDecl); isDecl && decl.Body != nil {
			decls[pass.TypesInfo.Defs[decl.Name]] = decl
			check(decl.Body, false)
		}
	})

	// what decides whether a change is shown runs while planning too, and is handed the plan's config as a ResourceData
	insp.Preorder([]ast.Node{(*ast.KeyValueExpr)(nil)}, func(n ast.Node) {
		kv, isPair := n.(*ast.KeyValueExpr)
		if !isPair || types.ExprString(kv.Key) != "DiffSuppressFunc" {
			return
		}

		switch suppress := ast.Unparen(kv.Value).(type) {
		case *ast.FuncLit:
			check(suppress.Body, true)

		case *ast.Ident:
			if decl := decls[pass.TypesInfo.Uses[suppress]]; decl != nil {
				check(decl.Body, true)
			}
		}
	})

	return nil, nil
}
