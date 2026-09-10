// Package AZR010 defines an analyzer that reports flatten* calls whose pointer argument is
// nil-checked by the caller: the flatten function should handle nil input itself, so every
// call site can pass the field straight through.
package AZR010

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/facts"
	"github.com/katbyte/azproviderlint/lib/nilguard"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks calls to flatten* functions (any package, case-insensitive prefix) that pass
// a pointer-typed variable or field chain from inside an if whose condition proves that chain
// non-nil — `if v.Ids != nil { out = flattenIds(v.Ids) }` — when the branch uses the chain for
// nothing else, so the check exists only for the flatten call. The nil handling belongs in
// the flatten function: one guard there covers every caller, and callers read as plain
// assignments. When the callee already returns early on a nil parameter (recorded per
// function by the nilinput fact, delegating wrappers included), the report says so and, for
// a bare `if x != nil { <one statement> }`, offers to drop the if.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR010",
	Doc:      "check that flatten functions handle nil input themselves instead of relying on nil checks at call sites",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR010_flatten_handles_nil_input/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer, nilInputAnalyzer},
	Run:      run,
}

// nilInputAnalyzer records, for every function with pointer parameters, which of them the
// body never dereferences unguarded: each use of the parameter is a nil comparison, a store
// (returned, assigned to a field, placed in a literal), an argument to a function known to
// handle nil in that position (a delegating `return other(input)`), or sits where the guard
// engine proves it non-nil (`if input == nil { return }`, `if input != nil { ... }`).
var nilInputAnalyzer = &analysis.Analyzer{
	Name:       "AZR010_nilinput",
	Doc:        "record which pointer parameters a function handles when nil",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	FactTypes:  []analysis.Fact{(*nilInput)(nil)},
	ResultType: reflect.TypeFor[*nilInputSet](),
}

func init() { nilInputAnalyzer.Run = runNilInput }

// nilInput is the fact: bit i set means parameter i is handled when nil.
type nilInput struct{ Mask uint64 }

type nilInputSet = facts.Set[nilInput, *nilInput]

func (*nilInput) AFact() {}

func (f *nilInput) String() string { return "nilinput(" + facts.Bits(f.Mask) + ")" }

// nilInputParam is one pointer parameter: its position and guard-engine key.
type nilInputParam struct {
	index int
	key   string
}

// nilInputCandidate is a declaration with at least one pointer parameter.
type nilInputCandidate struct {
	body    *ast.BlockStmt
	parents map[ast.Node]ast.Node
	params  map[types.Object]nilInputParam
}

func runNilInput(pass *analysis.Pass) (any, error) {
	return facts.Run(pass, nilInputAnalyzer,
		func(_ *types.Func, decl *ast.FuncDecl) (nilInputCandidate, bool) {
			c := nilInputCandidate{body: decl.Body, params: map[types.Object]nilInputParam{}}
			i := 0
			for _, field := range decl.Type.Params.List {
				for _, name := range field.Names {
					obj := pass.TypesInfo.Defs[name]
					if obj != nil {
						if _, ok := obj.Type().Underlying().(*types.Pointer); ok {
							if key, ok := nilguard.PathKey(pass, name); ok {
								c.params[obj] = nilInputParam{index: i, key: key}
							}
						}
					}
					i++
				}
			}
			if len(c.params) > 0 {
				c.parents = nilguard.ParentMap(decl.Body)
			}
			return c, len(c.params) > 0
		},
		func(s *nilInputSet, c nilInputCandidate) nilInput {
			var f nilInput
			for obj, param := range c.params {
				safe := true
				ast.Inspect(c.body, func(x ast.Node) bool {
					if id, ok := x.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj && !safeUse(pass, s, c.parents, id, param.key) {
						safe = false
					}
					return safe
				})
				if safe {
					f.Mask |= 1 << param.index
				}
			}
			return f
		})
}

// safeUse reports whether this use of a pointer parameter cannot dereference a nil: compared
// with nil, stored or returned, passed to a callee that handles nil there, or otherwise
// proven non-nil by the guard engine. An alias (`v := input`) is not followed, so it counts
// only where the assignment itself is guarded.
func safeUse(pass *analysis.Pass, s *nilInputSet, parents map[ast.Node]ast.Node, id *ast.Ident, key string) bool {
	var node ast.Node = id
	p := parents[node]
	for {
		paren, ok := p.(*ast.ParenExpr)
		if !ok {
			break
		}
		node, p = paren, parents[paren]
	}
	switch pp := p.(type) {
	case *ast.BinaryExpr:
		if (pp.Op == token.EQL || pp.Op == token.NEQ) && (astx.IsNilValue(pass, pp.X) || astx.IsNilValue(pass, pp.Y)) {
			return true
		}
	case *ast.CallExpr:
		mask := s.Get(astx.CalledFunc(pass, pp)).Mask
		for j, a := range pp.Args {
			if a == node && mask&(1<<j) != 0 {
				return true
			}
		}
	case *ast.AssignStmt:
		stored := true
		for _, lhs := range pp.Lhs {
			if _, isIdent := ast.Unparen(lhs).(*ast.Ident); isIdent {
				stored = false // an alias or a reassignment of the parameter: judged by the guard
			}
		}
		if stored {
			return true
		}
	case *ast.ReturnStmt, *ast.KeyValueExpr, *ast.CompositeLit:
		return true
	}
	return nilguard.Guarded(pass, parents, node, key)
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}
	handled := facts.Of[nilInput](pass, nilInputAnalyzer)

	nilguard.ForEachFunc(pass, insp, true, func(body *ast.BlockStmt, _ map[types.Object]bool, parents map[ast.Node]ast.Node) {
		ast.Inspect(body, func(x ast.Node) bool {
			call, ok := x.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn := astx.CalledFunc(pass, call)
			if fn == nil || !isFlatten(fn.Name()) {
				return true
			}
			for i, arg := range call.Args {
				t := pass.TypesInfo.TypeOf(arg)
				if t == nil {
					continue
				}
				if _, isPtr := t.Underlying().(*types.Pointer); !isPtr {
					continue
				}
				key, ok := nilguard.PathKey(pass, arg)
				if !ok {
					continue
				}
				ifs, keys := guardingIf(pass, parents, call, key)
				if ifs == nil || otherUse(pass, parents, ifs, keys) {
					continue
				}
				src := types.ExprString(arg)
				msg := "`" + fn.Name() + "(" + src + ")` is called under a nil check on `" + src + "` - handle nil inside the flatten function instead"
				var fixes []analysis.SuggestedFix
				if handled.Get(fn).Mask&(1<<i) != 0 {
					msg = "`" + fn.Name() + "` already handles a nil `" + src + "` - drop the nil check"
					fixes = dropCheck(pass, ifs)
				}
				pass.Report(analysis.Diagnostic{Pos: call.Pos(), Message: msg, SuggestedFixes: fixes})
			}
			return true
		})
	})
	return nil, nil
}

func isFlatten(name string) bool {
	return len(name) >= 7 && strings.EqualFold(name[:7], "flatten")
}

// guardingIf climbs from call to the nearest enclosing if whose condition proves key non-nil
// where the call sits: `key != nil` (in an && chain) for the body, `key == nil` (in an ||
// chain) for the else branch. An `x := chain` init makes x and chain interchangeable. It
// returns the if and the keys it proves.
func guardingIf(pass *analysis.Pass, parents map[ast.Node]ast.Node, call ast.Node, key string) (ifs *ast.IfStmt, keys []string) {
	child := call
	for node := parents[call]; node != nil; child, node = node, parents[node] {
		ifs, ok := node.(*ast.IfStmt)
		if !ok {
			continue
		}
		keys = []string{key}
		if init, ok := ifs.Init.(*ast.AssignStmt); ok && len(init.Lhs) == 1 && len(init.Rhs) == 1 {
			lk, lhsOK := nilguard.PathKey(pass, init.Lhs[0])
			rk, rhsOK := nilguard.PathKey(pass, init.Rhs[0])
			switch {
			case lhsOK && rhsOK && lk == key:
				keys = append(keys, rk)
			case lhsOK && rhsOK && rk == key:
				keys = append(keys, lk)
			}
		}
		if (ifs.Body == child && nilCompare(pass, ifs.Cond, keys, token.NEQ, token.LAND)) ||
			(ifs.Else == child && nilCompare(pass, ifs.Cond, keys, token.EQL, token.LOR)) {
			return ifs, keys
		}
	}
	return nil, nil
}

// nilCompare reports whether cond is `k <cmp> nil` for a k in keys, or a <join> chain
// containing one.
func nilCompare(pass *analysis.Pass, cond ast.Expr, keys []string, cmp, join token.Token) bool {
	c, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if c.Op == join {
		return nilCompare(pass, c.X, keys, cmp, join) || nilCompare(pass, c.Y, keys, cmp, join)
	}
	if c.Op != cmp {
		return false
	}
	other := c.X
	if astx.IsNilValue(pass, c.X) {
		other = c.Y
	} else if !astx.IsNilValue(pass, c.Y) {
		return false
	}
	k, ok := nilguard.PathKey(pass, other)
	return ok && slices.Contains(keys, k)
}

// otherUse reports whether either branch of ifs uses one of keys other than as a direct
// argument to a flatten* call — a field select, dereference, or any other call — which means
// the check is needed regardless.
func otherUse(pass *analysis.Pass, parents map[ast.Node]ast.Node, ifs *ast.IfStmt, keys []string) bool {
	found := false
	check := func(n ast.Node) {
		if n == nil || found {
			return
		}
		ast.Inspect(n, func(x ast.Node) bool {
			e, ok := x.(ast.Expr)
			if !ok || found {
				return !found
			}
			k, ok := nilguard.PathKey(pass, e)
			if !ok {
				return true
			}
			if !slices.Contains(keys, k) {
				return true
			}
			p := parents[x]
			for {
				paren, ok := p.(*ast.ParenExpr)
				if !ok {
					break
				}
				p = parents[paren]
			}
			if call, ok := p.(*ast.CallExpr); ok {
				if fn := astx.CalledFunc(pass, call); fn != nil && isFlatten(fn.Name()) {
					for _, a := range call.Args {
						if ast.Unparen(a) == x {
							return false // the flatten argument itself
						}
					}
				}
			}
			found = true
			return false
		})
	}
	check(ifs.Body)
	check(ifs.Else)
	return found
}

// dropCheck offers to unwrap `if x != nil { <stmt> }` — no init, no else, a plain `!= nil`
// condition, a single-statement body — by deleting the if line and the closing brace line and
// unindenting the statement, so trailing comments stay put.
func dropCheck(pass *analysis.Pass, ifs *ast.IfStmt) []analysis.SuggestedFix {
	cmp, ok := ast.Unparen(ifs.Cond).(*ast.BinaryExpr)
	if !ok || cmp.Op != token.NEQ || ifs.Init != nil || ifs.Else != nil || len(ifs.Body.List) != 1 {
		return nil
	}
	stmt := ifs.Body.List[0]
	src, ok := astx.SourceText(pass, ifs)
	if !ok {
		return nil
	}
	rbrace := int(ifs.Body.Rbrace - ifs.Pos())
	lineStart := strings.LastIndex(string(src[:rbrace]), "\n")
	if lineStart < 0 {
		return nil // single-line if: nothing to unwrap cleanly
	}
	edits := []analysis.TextEdit{
		{Pos: ifs.Pos(), End: stmt.Pos()},
		{Pos: ifs.Pos() + token.Pos(lineStart), End: ifs.Body.Rbrace + 1},
	}
	// continuation lines of the statement lose one level of indentation
	stmtSrc := string(src)[int(stmt.Pos()-ifs.Pos()):int(stmt.End()-ifs.Pos())]
	for off := strings.Index(stmtSrc, "\n\t"); off >= 0; {
		at := stmt.Pos() + token.Pos(off) + 1
		edits = append(edits, analysis.TextEdit{Pos: at, End: at + 1})
		next := strings.Index(stmtSrc[off+2:], "\n\t")
		if next < 0 {
			break
		}
		off += 2 + next
	}
	return []analysis.SuggestedFix{{Message: "Drop the nil check", TextEdits: edits}}
}
