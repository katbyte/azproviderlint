// Writes: what an assignment does to a path — proves it, aliases it, or leaves it unknown —
// including writes buried inside compound statements.
package nilguard

import (
	"go/ast"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
)

// nestedWrite is one write found inside a statement: the path written and, for a one-to-one
// assignment, the value; nil when the value is unknown (a range variable, inc/dec, a
// multi-result call).
type nestedWrite struct {
	key   string
	value ast.Expr
	at    ast.Node // the writing statement, where value is judged
}

// nestedWrites returns every assignment, inc/dec, or range-variable target anywhere inside
// n, in no particular order.
func nestedWrites(pass *analysis.Pass, n ast.Node) []nestedWrite {
	var writes []nestedWrite
	add := func(e, value ast.Expr, at ast.Node) {
		if e == nil {
			return
		}
		if k, ok := PathKey(pass, e); ok {
			writes = append(writes, nestedWrite{key: k, value: value, at: at})
		}
	}
	ast.Inspect(n, func(x ast.Node) bool {
		switch s := x.(type) {
		case *ast.AssignStmt:
			for i, lhs := range s.Lhs {
				var value ast.Expr
				if len(s.Lhs) == len(s.Rhs) {
					value = s.Rhs[i]
				}
				add(lhs, value, s)
			}
		case *ast.IncDecStmt:
			add(s.X, nil, s)
		case *ast.RangeStmt:
			add(s.Key, nil, s)
			add(s.Value, nil, s)
		}
		return true
	})
	return writes
}

// writesPath reports whether any write anywhere inside n touches key — the raw question, used
// for identity: an alias or companion recorded before such a write no longer describes key.
// assignsPath is the nil-safety variant that forgives writes proven non-nil.
func writesPath(pass *analysis.Pass, n ast.Node, key string) bool {
	for _, w := range nestedWrites(pass, n) {
		if touches(w.key, key) {
			return true
		}
	}
	return false
}

// assignsPath reports whether any write anywhere inside n could leave key nil: a write to key
// from anything but a value proven non-nil at that point, or a write to one of its prefixes.
// nil n never does; depth is the guard engine's recursion budget, shared with it so the two
// cannot bounce off each other indefinitely.
func assignsPath(pass *analysis.Pass, parents map[ast.Node]ast.Node, n ast.Node, key string, depth int) bool {
	if n == nil {
		return false
	}
	for _, w := range nestedWrites(pass, n) {
		if !touches(w.key, key) {
			continue
		}
		// judge the write where it happens: is key still proven non-nil at the end of the
		// block it sits in? That applies the ordinary rules — non-nil source, checked
		// companion, alias of a guarded chain, a fresh nil check after a re-fetch of a
		// prefix — to the nested statement and what follows it in its block
		if depth < 4 {
			block := w.at
			for parents[block] != nil && stmtList(parents[block]) == nil {
				block = parents[block]
			}
			if stmts := stmtList(parents[block]); stmts != nil {
				switch guarded, alias, _ := precededByGuard(pass, parents, stmts, nil, key, depth+1); {
				case guarded:
					continue
				case alias != "" && guardedKey(pass, parents, stmts[len(stmts)-1], alias, depth+1):
					continue
				}
			}
		}
		return true
	}
	return false
}

// endsAssignedNonNil reports whether the last write to key among stmts is a top-level
// assignment of a provably non-nil value — so key is non-nil once they have run.
func endsAssignedNonNil(pass *analysis.Pass, parents map[ast.Node]ast.Node, stmts []ast.Stmt, key string, depth int) bool {
	nonNil := false
	for _, s := range stmts {
		if assign, ok := s.(*ast.AssignStmt); ok && len(assign.Lhs) == len(assign.Rhs) {
			direct := false
			for i, lhs := range assign.Lhs {
				if lk, ok := PathKey(pass, lhs); ok && lk == key {
					nonNil, direct = isNonNilSource(pass, assign.Rhs[i]), true
				}
			}
			if direct {
				continue
			}
		}
		if assignsPath(pass, parents, s, key, depth) {
			nonNil = false
		}
	}
	return nonNil
}

// assignmentGuard classifies what assign does to key; matched is false when it does not touch
// key or a prefix of it. A provable source proves the guard, an alias of another chain
// redirects the search, a multi-result call is guarded when its error/ok companion was
// checked in following (the statements up to the dereference, or the if the call is the init
// of), and anything else settles the value as unknown.
func assignmentGuard(pass *analysis.Pass, assign *ast.AssignStmt, key string, following []ast.Stmt) (guarded bool, alias string, settled, matched bool) {
	for j, lhs := range assign.Lhs {
		lk, ok := PathKey(pass, lhs)
		if !ok || !touches(lk, key) {
			continue
		}
		if len(assign.Lhs) == len(assign.Rhs) {
			if lk == key && isNonNilSource(pass, assign.Rhs[j]) {
				return true, "", false, true
			}
			if rk, ok := PathKey(pass, assign.Rhs[j]); ok {
				return false, rk + strings.TrimPrefix(key, lk), false, true
			}
			if lk != key {
				// the field's value is whatever the literal gave it: `x := T{F: pointer.To(v)}`
				// proves x.F, `x := T{F: y}` makes x.F an alias of y; `id :=
				// commonids.NewCompositeResourceID(a, b)` likewise stores its pointer
				// arguments as First/Second
				src := literalField(assign.Rhs[j], strings.TrimPrefix(key, lk))
				if src == nil {
					src = compositeIDArg(pass, assign.Rhs[j], strings.TrimPrefix(key, lk))
				}
				if src != nil {
					if isNonNilSource(pass, src) {
						return true, "", false, true
					}
					if sk, ok := PathKey(pass, src); ok {
						return false, sk, false, true
					}
				}
			}
			return false, "", true, true
		}
		if len(assign.Rhs) == 1 {
			if call, isCall := ast.Unparen(assign.Rhs[0]).(*ast.CallExpr); isCall {
				// `x, err := f()` / `x, ok := f()` — x is valid when f is known never to
				// return nil there, or once a later `if err != nil { return }` /
				// `if !ok { return }` exited (the Go contract)
				if lk == key {
					if nonNilResult(pass, astx.CalledFunc(pass, call), j) {
						return true, "", false, true
					}
					return companionCheckedBetween(pass, assign, following), "", true, true
				}
				// `id, err := commonids.ParseCompositeResourceID(s, a, b)` likewise stores a
				// and b as First/Second
				if arg := compositeIDArg(pass, assign.Rhs[0], strings.TrimPrefix(key, lk)); arg != nil {
					if isNonNilSource(pass, arg) {
						return companionCheckedBetween(pass, assign, following), "", true, true
					}
					if ak, ok := PathKey(pass, arg); ok {
						return false, ak, false, true
					}
				}
			}
		}
		return false, "", true, true
	}
	return false, "", false, false
}
