// Package AZG013 defines an analyzer that reports functions that always call each other, so that none of them can return.
package AZG013

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/astx"
)

// Analyzer checks for a loop of two or more functions in one package where each calls the next on every path through it. Calling any of them
// never returns: the program runs until it is out of stack. A function that always calls itself is left to staticcheck's SA5007.
var Analyzer = &analysis.Analyzer{
	Name: "AZG013",
	Doc:  "check for functions that always call each other and so never return",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG013_functions_always_call_each_other/README.md",
	Run:  run,
}

// calls that end the function they are made from, so nothing after them is sure to run
var noReturn = map[string]bool{"Exit": true, "Fatal": true, "Fatalf": true, "Fatalln": true, "FailNow": true, "Skip": true, "Skipf": true, "SkipNow": true, "Goexit": true, "Panic": true, "Panicf": true, "Panicln": true}

func run(pass *analysis.Pass) (any, error) {
	// every function declared here, in the order of the files and then of the lines
	var order []*types.Func
	decls := map[*types.Func]*ast.FuncDecl{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}

			if fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func); ok {
				order = append(order, fn)
				decls[fn] = fd
			}
		}
	}

	// for each, the functions declared here that it calls whatever path is taken, with the first such call
	always := map[*types.Func]map[*types.Func]*ast.CallExpr{}
	for _, fn := range order {
		always[fn] = map[*types.Func]*ast.CallExpr{}
		for _, call := range sureCalls(pass, decls[fn].Body) {
			callee := astx.CalledFunc(pass, call)
			if callee == nil {
				continue
			}

			callee = callee.Origin()
			if decls[callee] != nil && callee != fn && always[fn][callee] == nil {
				always[fn][callee] = call
			}
		}
	}

	// the shortest loop back to each function, reported from the first function in it so that each loop is reported once
	reported := map[*types.Func]bool{}
	for _, fn := range order {
		if reported[fn] {
			continue
		}

		loop := shortestLoop(fn, order, always)
		if loop == nil {
			continue
		}

		for _, member := range loop {
			reported[member] = true
		}

		var chain strings.Builder
		chain.WriteString(name(loop[0]))
		for i, member := range loop[1:] {
			if i == 0 {
				chain.WriteString(" always calls " + name(member))
				continue
			}

			chain.WriteString(", which always calls " + name(member))
		}

		neither := "neither returns"
		if len(loop) > 2 {
			neither = "none of them return"
		}

		pass.Reportf(always[fn][loop[1]].Pos(), "%s, which always calls %s again, so %s", chain.String(), name(fn), neither)
	}

	return nil, nil
}

// shortestLoop returns the fewest functions that lead from start back to start by calls that are always made, start first, or nil when there
// is no way back.
func shortestLoop(start *types.Func, order []*types.Func, always map[*types.Func]map[*types.Func]*ast.CallExpr) []*types.Func {
	from := map[*types.Func]*types.Func{}
	queue := []*types.Func{start}
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]

		// in declaration order, so the same loop is found on every run
		for _, callee := range order {
			if always[fn][callee] == nil {
				continue
			}

			if callee == start {
				loop := []*types.Func{fn}
				for fn != start {
					fn = from[fn]
					loop = append(loop, fn)
				}

				slices.Reverse(loop)

				return loop
			}

			if from[callee] == nil {
				from[callee] = fn
				queue = append(queue, callee)
			}
		}
	}

	return nil
}

// sureCalls returns the calls a block makes on every path through it, in order: those in its statements up to the first one that can branch,
// return or end the function.
func sureCalls(pass *analysis.Pass, block *ast.BlockStmt) []*ast.CallExpr {
	var calls []*ast.CallExpr
	for _, stmt := range block.List {
		found, carryOn := stmtCalls(pass, stmt)
		calls = append(calls, found...)
		if !carryOn {
			break
		}
	}

	return calls
}

// stmtCalls returns the calls a statement is sure to make, and whether the statement after it is sure to run.
func stmtCalls(pass *analysis.Pass, stmt ast.Stmt) (calls []*ast.CallExpr, carryOn bool) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		calls = exprCalls(s.X)
	case *ast.AssignStmt:
		calls = exprCalls(append(slices.Clone(s.Rhs), s.Lhs...)...)
	case *ast.IncDecStmt:
		calls = exprCalls(s.X)
	case *ast.SendStmt:
		calls = exprCalls(s.Chan, s.Value)
	case *ast.DeclStmt:
		ast.Inspect(s, func(n ast.Node) bool {
			if spec, ok := n.(*ast.ValueSpec); ok {
				calls = append(calls, exprCalls(spec.Values...)...)
			}

			return true
		})
	case *ast.BlockStmt:
		calls = sureCalls(pass, s)

		// a block that stopped early stops what holds it too
		return calls, endsWell(pass, s)
	case *ast.ReturnStmt:
		return exprCalls(s.Results...), false
	case *ast.IfStmt:
		// only what is asked before the branch is sure to run
		return append(initCalls(pass, s.Init), exprCalls(s.Cond)...), false
	case *ast.SwitchStmt:
		return append(initCalls(pass, s.Init), exprCalls(s.Tag)...), false
	case *ast.ForStmt:
		return append(initCalls(pass, s.Init), exprCalls(s.Cond)...), false
	case *ast.RangeStmt:
		return exprCalls(s.X), false
	case *ast.DeferStmt, *ast.GoStmt, *ast.EmptyStmt:
		// run later or elsewhere, and what follows still runs
	default:
		return nil, false
	}

	// a call that ends the function means nothing after it is sure to run
	return calls, !slices.ContainsFunc(calls, func(call *ast.CallExpr) bool { return ends(pass, call) })
}

// endsWell reports whether every statement of a block lets the next one run.
func endsWell(pass *analysis.Pass, block *ast.BlockStmt) bool {
	for _, stmt := range block.List {
		if _, carryOn := stmtCalls(pass, stmt); !carryOn {
			return false
		}
	}

	return true
}

func initCalls(pass *analysis.Pass, init ast.Stmt) []*ast.CallExpr {
	if init == nil {
		return nil
	}

	calls, _ := stmtCalls(pass, init)

	return calls
}

// exprCalls returns the calls that working out the expressions is sure to make. The right side of && and || may be skipped, and the body of a
// function written in place only runs if something calls it.
func exprCalls(exprs ...ast.Expr) []*ast.CallExpr {
	var calls []*ast.CallExpr
	for _, expr := range exprs {
		if expr == nil {
			continue
		}

		ast.Inspect(expr, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncLit:
				return false
			case *ast.BinaryExpr:
				if n.Op == token.LAND || n.Op == token.LOR {
					calls = append(calls, exprCalls(n.X)...)
					return false
				}
			case *ast.CallExpr:
				calls = append(calls, n)
			}

			return true
		})
	}

	return calls
}

// ends reports whether a call stops the function it is made from: panic, or one of the usual ways to exit, fail or skip.
func ends(pass *analysis.Pass, call *ast.CallExpr) bool {
	if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok {
		if _, builtin := pass.TypesInfo.Uses[id].(*types.Builtin); builtin && id.Name == "panic" {
			return true
		}
	}

	fn := astx.CalledFunc(pass, call)

	return fn != nil && fn.Pkg() != pass.Pkg && noReturn[fn.Name()]
}

// name is how a report names a function: its name, with the type it belongs to in front when it is a method.
func name(fn *types.Func) string {
	recv := fn.Signature().Recv()
	if recv == nil {
		return fn.Name()
	}

	t := recv.Type()
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	if named, ok := types.Unalias(t).(*types.Named); ok {
		return fmt.Sprintf("%s.%s", named.Obj().Name(), fn.Name())
	}

	return fn.Name()
}
