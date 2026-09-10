// The guard walk: from a dereference outward through enclosing conditions and preceding
// statements, deciding whether a path is proven non-nil where it is used.

package nilguard

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Guarded walks at's ancestors looking for a condition, early-exit, or non-nil assignment
// that proves key is not nil where at sits. When the walk finds that key aliases another chain
// (`payload := existing.Model`), it restarts with the source chain so guards on either side
// of the assignment count; the depth cap bounds alias-of-alias chains.
func Guarded(pass *analysis.Pass, parents map[ast.Node]ast.Node, at ast.Node, key string) bool {
	return guardedKey(pass, parents, at, key, 0)
}

func guardedKey(pass *analysis.Pass, parents map[ast.Node]ast.Node, at ast.Node, key string, depth int) bool {
	if depth > 4 {
		return false
	}
	child := at
	for node := parents[at]; node != nil; child, node = node, parents[node] {
		switch p := node.(type) {
		case *ast.BinaryExpr:
			// short-circuit: in `x != nil && *x...` / `x == nil || *x...` the right operand
			// only evaluates once the left proved x non-nil
			if p.Y == child && ((p.Op == token.LAND && impliesNonNil(pass, p.X, key)) ||
				(p.Op == token.LOR && impliedByNil(pass, p.X, key))) {
				return true
			}
		case *ast.IfStmt:
			if p.Body == child || p.Else == child {
				// a nil check on an alias of the chain (`if v := x.F; v != nil { *x.F }`,
				// or `v := x.F` earlier in the block) covers the chain itself
				for _, k := range equivalentKeys(pass, parents, p, key) {
					if p.Body == child && impliesNonNil(pass, p.Cond, k) {
						return true
					}
					// the else branch runs when the condition is false; a pure-||
					// condition with an `x == nil` disjunct being false proves x non-nil
					if p.Else == child && impliedByNil(pass, p.Cond, k) {
						return true
					}
				}
				// `if x, ok := f(); ok { *x }`, `x, err := f(); if err == nil { *x }`, and
				// the else branch of `if err != nil` / `if !ok`
				errKey, okKey := companionOf(pass, parents, p, key)
				if p.Body == child && condProvesValid(pass, p.Cond, errKey, okKey) {
					return true
				}
				if p.Else == child && condProvesInvalid(pass, p.Cond, errKey, okKey) {
					return true
				}
			}
			// an init that assigns key settles it like any other assignment — for the
			// condition too, which runs right after it
			if init, ok := p.Init.(*ast.AssignStmt); ok && (p.Cond == child || p.Body == child || p.Else == child) {
				switch guarded, alias, settled, matched := assignmentGuard(pass, init, key, nil); {
				case !matched:
				case guarded:
					return true
				case alias != "":
					return guardedKey(pass, parents, at, alias, depth+1)
				case settled:
					return false
				}
			}
		case *ast.ForStmt:
			if p.Body == child && p.Cond != nil && impliesNonNil(pass, p.Cond, key) {
				return true // re-checked every iteration, so writes in the body are fine
			}
			// otherwise a write anywhere in the body reaches the dereference on the next
			// iteration, before any guard outside the loop
			if p.Body == child && assignsPath(pass, parents, p.Body, key, depth) {
				return false
			}
		case *ast.RangeStmt:
			if p.Body == child && assignsPath(pass, parents, p.Body, key, depth) {
				return false
			}
		case *ast.CaseClause:
			// the clause fires when any listed expression is true, so all must prove it
			if len(p.List) > 0 && slices.ContainsFunc(p.Body, func(s ast.Stmt) bool { return s == child }) {
				all := true
				for _, ce := range p.List {
					all = all && impliesNonNil(pass, ce, key)
				}
				if all {
					return true
				}
			}
		}
		stmts := stmtList(node)
		if stmts == nil {
			continue
		}
		switch ok, alias, settled := precededByGuard(pass, parents, stmts, child, key, depth); {
		case ok:
			return true
		case alias != "":
			return guardedKey(pass, parents, at, alias, depth+1)
		case settled:
			return false // an assignment made the value unknown; guards further out are stale
		}
	}
	return false
}

// touches reports whether a write to the path written covers key: the same path, or a prefix
// of it (writing `x` replaces `x.F` too).
func touches(written, key string) bool {
	return written == key || strings.HasPrefix(key, written+".")
}

// indexOf returns child's position in stmts, or len(stmts) when child is nil or absent — the
// point after all of them.
func indexOf(stmts []ast.Stmt, child ast.Node) int {
	for i, s := range stmts {
		if s == child {
			return i
		}
	}
	return len(stmts)
}

// stmtList returns the statement list a block, case clause, or comm clause holds, nil for any
// other node.
func stmtList(n ast.Node) []ast.Stmt {
	switch p := n.(type) {
	case *ast.BlockStmt:
		return p.List
	case *ast.CaseClause:
		return p.Body
	case *ast.CommClause:
		return p.Body
	}
	return nil
}

// precedingStmts visits, nearest first, every statement that executed before at on the way
// to it: at's own init when it is an if, then the statements before it in each enclosing
// statement list, and the init of each enclosing if/for/switch. visit returning false stops
// the walk.
func precedingStmts(parents map[ast.Node]ast.Node, at ast.Node, visit func(ast.Stmt) bool) {
	if ifs, ok := at.(*ast.IfStmt); ok && ifs.Init != nil && !visit(ifs.Init) {
		return
	}
	child := at
	for node := parents[child]; node != nil; child, node = node, parents[node] {
		var init ast.Stmt
		switch p := node.(type) {
		case *ast.IfStmt:
			init = p.Init
		case *ast.ForStmt:
			init = p.Init
			// the whole body ran before this iteration: its writes count as preceding
			if p.Body == child && !visit(p.Body) {
				return
			}
		case *ast.RangeStmt:
			if p.Body == child && !visit(p.Body) {
				return
			}
		case *ast.SwitchStmt:
			init = p.Init
		case *ast.TypeSwitchStmt:
			init = p.Init
		}
		if init != nil && !visit(init) {
			return
		}
		stmts := stmtList(node)
		for i := indexOf(stmts, child) - 1; i >= 0; i-- {
			if !visit(stmts[i]) {
				return
			}
		}
	}
}

// equivalentKeys returns key followed by the keys of chains holding the same value where at
// executes: variables assigned from key's chain before at (`v := x.F` — in a block, or an
// enclosing if's init) and, once the walk reaches the assignment that defined key itself, the
// chain it was copied from (`x := y.F` makes `x.G` equivalent to `y.F.G`). A prefix alias
// carries the remaining path. The walk stops at key's own assignment — anything earlier
// aliases a stale value — and a candidate written between its definition and at is dropped.
func equivalentKeys(pass *analysis.Pass, parents map[ast.Node]ast.Node, at ast.Node, key string) []string {
	keys := []string{key}
	var seen []ast.Stmt // statements between the one being examined and at
	stale := func(k string) bool {
		return slices.ContainsFunc(seen, func(s ast.Stmt) bool { return writesPath(pass, s, k) })
	}
	precedingStmts(parents, at, func(s ast.Stmt) bool {
		defer func() { seen = append(seen, s) }()
		assign, ok := s.(*ast.AssignStmt)
		if !ok {
			return !writesPath(pass, s, key) // a compound statement writing key ends the search
		}
		for i, lhs := range assign.Lhs {
			lk, ok := PathKey(pass, lhs)
			if !ok {
				continue
			}
			var rk string
			var rkOK bool
			if len(assign.Lhs) == len(assign.Rhs) {
				rk, rkOK = PathKey(pass, assign.Rhs[i])
			}
			if touches(lk, key) {
				if k := rk + strings.TrimPrefix(key, lk); rkOK && !stale(k) {
					keys = append(keys, k)
				}
				return false
			}
			if k := lk + strings.TrimPrefix(key, rk); rkOK && touches(rk, key) && !stale(k) {
				keys = append(keys, k)
			}
		}
		return true
	})
	return keys
}

// companionOf returns the error and ok-bool keys returned alongside key by the nearest
// preceding `key, err := f()` / `key, ok := f()`, or empty strings when key's latest
// assignment is not such a call or the companion was written again before at.
func companionOf(pass *analysis.Pass, parents map[ast.Node]ast.Node, at ast.Node, key string) (errKey, okKey string) {
	var seen []ast.Stmt
	precedingStmts(parents, at, func(s ast.Stmt) bool {
		defer func() { seen = append(seen, s) }()
		assign, ok := s.(*ast.AssignStmt)
		if !ok {
			return !writesPath(pass, s, key)
		}
		for _, lhs := range assign.Lhs {
			lk, ok := PathKey(pass, lhs)
			if !ok || !touches(lk, key) {
				continue
			}
			// only a call's extra results are companions: a type assertion's or map
			// lookup's ok says nothing about the value being nil
			if _, isCall := ast.Unparen(assign.Rhs[0]).(*ast.CallExpr); isCall && lk == key && len(assign.Lhs) > 1 && len(assign.Rhs) == 1 {
				errKey, okKey = companionKeys(pass, assign)
				written := func(k string) bool {
					return k != "" && slices.ContainsFunc(seen, func(s ast.Stmt) bool { return writesPath(pass, s, k) })
				}
				if written(errKey) {
					errKey = ""
				}
				if written(okKey) {
					okKey = ""
				}
			}
			return false
		}
		return true
	})
	return errKey, okKey
}

// precededByGuard reports whether a statement before child in stmts (all of them, when child
// is nil) proves key non-nil: an
// `if key == nil { <terminating> }` early exit, an `if key == nil { key = &T{} }` default
// init, an assignment of a provably non-nil value, or a multi-result call whose companion
// error/ok result was checked before child — as a following statement or as the if the call
// is the init of. When key was assigned from another chain (`payload := existing.Model`), the
// source chain is returned as alias so the caller can restart the guard search with it. Any
// other assignment settles the value as unknown: settled tells the caller to stop — guards in
// enclosing scopes predate the assignment and no longer hold.
func precededByGuard(pass *analysis.Pass, parents map[ast.Node]ast.Node, stmts []ast.Stmt, child ast.Node, key string, depth int) (guarded bool, alias string, settled bool) {
	idx := indexOf(stmts, child) // a nil child asks about the value once all of stmts have run
	for i := idx - 1; i >= 0; i-- {
		switch s := stmts[i].(type) {
		case *ast.IfStmt:
			if terminates(s.Body) && !assignsPath(pass, parents, s.Else, key, depth) {
				// code after the if only runs when the condition was false (via a
				// non-terminating else, provided it left key alone)
				for _, k := range equivalentKeys(pass, parents, s, key) {
					if impliedByNil(pass, s.Cond, k) {
						return true, "", false
					}
				}
			} else if s.Else == nil && impliedByNil(pass, s.Cond, key) && endsAssignedNonNil(pass, parents, s.Body.List, key, depth) {
				return true, "", false
			}
			if init, ok := s.Init.(*ast.AssignStmt); ok {
				if guarded, alias, settled, matched := assignmentGuard(pass, init, key, stmts[i:idx]); matched {
					return guarded, alias, settled
				}
			}
		case *ast.AssignStmt:
			if guarded, alias, settled, matched := assignmentGuard(pass, s, key, stmts[i+1:idx]); matched {
				// the alias only stands while its source chain is untouched afterwards
				for _, between := range stmts[i+1 : idx] {
					if alias != "" && assignsPath(pass, parents, between, alias, depth) {
						return false, "", true
					}
				}
				return guarded, alias, settled
			}
			continue // an assignment to something else cannot hide a write to key
		}
		// any other write to key buried in the statement (`if c { x = nil }`, a loop, a
		// switch) leaves the value unknown
		if assignsPath(pass, parents, stmts[i], key, depth) {
			return false, "", true
		}
	}
	return false, "", false
}
