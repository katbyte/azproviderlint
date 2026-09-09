// Package nilguard is the shared guard engine behind AZG008 and AZG009: it decides whether a
// pointer-typed variable or selector chain is provably non-nil at a given point in a function
// body.
package nilguard

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/pointerpkg"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// ForEachFunc visits every function body in the pass exactly once — package-level function
// literals (var f = func() {...}) are reached through their own visit, nested literals through
// their enclosing FuncDecl, keeping guards outside a closure visible to dereferences inside
// it. Bodies in _test.go files are skipped when tests is false. visit receives the body, the
// function's (and receiver's) parameter objects, and a child-to-parent map covering the body.
func ForEachFunc(pass *analysis.Pass, insp *inspector.Inspector, tests bool, visit func(body *ast.BlockStmt, params map[types.Object]bool, parents map[ast.Node]ast.Node)) {
	var declRanges [][2]token.Pos
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		declRanges = append(declRanges, [2]token.Pos{n.Pos(), n.End()})
	})
	inDecl := func(pos token.Pos) bool {
		for _, r := range declRanges {
			if r[0] <= pos && pos < r[1] {
				return true
			}
		}
		return false
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(n ast.Node) {
		var body *ast.BlockStmt
		params := map[types.Object]bool{}
		collectParams := func(fl *ast.FieldList) {
			if fl == nil {
				return
			}
			for _, field := range fl.List {
				for _, name := range field.Names {
					if obj := pass.TypesInfo.Defs[name]; obj != nil {
						params[obj] = true
					}
				}
			}
		}
		switch fn := n.(type) {
		case *ast.FuncDecl:
			body = fn.Body
			collectParams(fn.Recv)
			collectParams(fn.Type.Params)
		case *ast.FuncLit:
			if inDecl(fn.Pos()) {
				return // visited via its enclosing FuncDecl
			}
			body = fn.Body
			collectParams(fn.Type.Params)
		}
		if body == nil {
			return
		}

		if !tests && strings.HasSuffix(pass.Fset.Position(body.Pos()).Filename, "_test.go") {
			return
		}

		visit(body, params, parentMap(body))
	})
}

// PathKey canonicalizes a variable or field chain (`props.Status`) so guard conditions and
// dereference operands can be compared; the root must resolve to a variable.
func PathKey(pass *analysis.Pass, e ast.Expr) (string, bool) {
	switch x := ast.Unparen(e).(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[x]
		if obj == nil {
			obj = pass.TypesInfo.Defs[x] // the ident being defined by `x := ...`
		}
		if v, ok := obj.(*types.Var); ok {
			return v.Id() + "@" + pass.Fset.Position(v.Pos()).String(), true
		}
	case *ast.SelectorExpr:
		if base, ok := PathKey(pass, x.X); ok {
			return base + "." + x.Sel.Name, true
		}
	}
	return "", false
}

// DerefNeedsPointer reports whether star's context needs the pointee to be addressable — so a
// pointer.From copy cannot stand in: an assignment or inc/dec target (`*x = v`, `(*x).F = v`,
// `(*x)[i] = v` on an array, `(*x)++`), the operand of & (`&*x`, `&(*x).F`), a slice of an
// array (`(*x)[:]`), or the receiver of a pointer method (`(*x).M()`). Field selections and
// array indexes propagate the requirement; a selection that itself goes through a pointer
// field, or an index into a slice or map, is addressable on its own.
func DerefNeedsPointer(pass *analysis.Pass, parents map[ast.Node]ast.Node, star *ast.StarExpr) bool {
	var outer ast.Node = star
	for {
		switch p := parents[outer].(type) {
		case *ast.ParenExpr:
			outer = p
		case *ast.SelectorExpr:
			sel := pass.TypesInfo.Selections[p]
			if p.X != outer || sel == nil || sel.Indirect() {
				return false
			}
			if sel.Kind() == types.MethodVal {
				_, ptrRecv := sel.Obj().(*types.Func).Type().(*types.Signature).Recv().Type().(*types.Pointer)
				return ptrRecv
			}
			outer = p
		case *ast.IndexExpr:
			if p.X != outer || !isArray(pass, p.X) {
				return false
			}
			outer = p
		case *ast.SliceExpr:
			return p.X == outer && isArray(pass, p.X)
		case *ast.AssignStmt:
			for _, lhs := range p.Lhs {
				if lhs == outer {
					return true
				}
			}
			return false
		case *ast.UnaryExpr:
			return p.Op == token.AND
		case *ast.IncDecStmt:
			return true
		default:
			return false
		}
	}
}

func isArray(pass *analysis.Pass, e ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(e)
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Array)
	return ok
}

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
			if p.Body != child && p.Else != child {
				continue
			}
			// a nil check on an alias of the chain (`if v := x.F; v != nil { *x.F }`, or
			// `v := x.F` earlier in the block) covers the chain itself
			for _, k := range equivalentKeys(pass, parents, p, key) {
				if p.Body == child && impliesNonNil(pass, p.Cond, k) {
					return true
				}
				// the else branch runs when the condition is false; a pure-|| condition
				// with an `x == nil` disjunct being false proves x non-nil
				if p.Else == child && impliedByNil(pass, p.Cond, k) {
					return true
				}
			}
			// `if x, ok := f(); ok { *x }`, `x, err := f(); if err == nil { *x }`, and the
			// else branch of `if err != nil` / `if !ok`
			errKey, okKey := companionOf(pass, parents, p, key)
			if p.Body == child && condProvesValid(pass, p.Cond, errKey, okKey) {
				return true
			}
			if p.Else == child && condProvesInvalid(pass, p.Cond, errKey, okKey) {
				return true
			}
			// an init that assigns key settles it like any other assignment
			if init, ok := p.Init.(*ast.AssignStmt); ok {
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
				return true
			}
		case *ast.CaseClause:
			// the clause fires when any listed expression is true, so all must prove it
			if len(p.List) > 0 && containsStmt(p.Body, child) {
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
		switch ok, alias, settled := precededByGuard(pass, parents, stmts, child, key); {
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
		case *ast.SwitchStmt:
			init = p.Init
		case *ast.TypeSwitchStmt:
			init = p.Init
		}
		if init != nil && !visit(init) {
			return
		}
		stmts := stmtList(node)
		idx := -1
		for i, s := range stmts {
			if s == child {
				idx = i
				break
			}
		}
		for i := idx - 1; i >= 0; i-- {
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
// aliases a stale value — and an alias reassigned after its definition is dropped.
func equivalentKeys(pass *analysis.Pass, parents map[ast.Node]ast.Node, at ast.Node, key string) []string {
	keys := []string{key}
	// chains assigned after the statement being examined (the walk runs backwards), so any
	// alias involving them, or a path under them, is stale
	var stale []string
	isStale := func(k string) bool {
		for _, sk := range stale {
			if k == sk || strings.HasPrefix(k, sk+".") {
				return true
			}
		}
		return false
	}
	precedingStmts(parents, at, func(s ast.Stmt) bool {
		assign, ok := s.(*ast.AssignStmt)
		if !ok {
			return true
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
			if lk == key || strings.HasPrefix(key, lk+".") {
				if rkOK && !isStale(rk) {
					keys = append(keys, rk+strings.TrimPrefix(key, lk))
				}
				return false
			}
			if rkOK && !isStale(lk) {
				if rk == key {
					keys = append(keys, lk)
				} else if strings.HasPrefix(key, rk+".") {
					keys = append(keys, lk+strings.TrimPrefix(key, rk))
				}
			}
			stale = append(stale, lk)
		}
		return true
	})
	return keys
}

// companionOf returns the error and ok-bool keys returned alongside key by the nearest
// preceding `key, err := f()` / `key, ok := f()`, or empty strings when key's latest
// assignment is not such a call.
func companionOf(pass *analysis.Pass, parents map[ast.Node]ast.Node, at ast.Node, key string) (errKey, okKey string) {
	reassigned := map[string]bool{} // companions overwritten after the call no longer speak for it
	precedingStmts(parents, at, func(s ast.Stmt) bool {
		assign, ok := s.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			lk, ok := PathKey(pass, lhs)
			if !ok {
				continue
			}
			if lk == key || strings.HasPrefix(key, lk+".") {
				// only a call's extra results are companions: a type assertion's or map
				// lookup's ok says nothing about the value being nil
				if _, isCall := ast.Unparen(assign.Rhs[0]).(*ast.CallExpr); isCall && lk == key && len(assign.Lhs) > 1 && len(assign.Rhs) == 1 {
					errKey, okKey = companionKeys(pass, assign)
					if reassigned[errKey] {
						errKey = ""
					}
					if reassigned[okKey] {
						okKey = ""
					}
				}
				return false
			}
			reassigned[lk] = true
		}
		return true
	})
	return errKey, okKey
}

func containsStmt(stmts []ast.Stmt, n ast.Node) bool {
	for _, s := range stmts {
		if s == n {
			return true
		}
	}
	return false
}

// precededByGuard reports whether a statement before child in stmts proves key non-nil: an
// `if key == nil { <terminating> }` early exit, an `if key == nil { key = &T{} }` default
// init, an assignment of a provably non-nil value, or a multi-result call whose companion
// error/ok result was checked before child — as a following statement or as the if the call
// is the init of. When key was assigned from another chain (`payload := existing.Model`), the
// source chain is returned as alias so the caller can restart the guard search with it. Any
// other assignment settles the value as unknown: settled tells the caller to stop — guards in
// enclosing scopes predate the assignment and no longer hold.
func precededByGuard(pass *analysis.Pass, parents map[ast.Node]ast.Node, stmts []ast.Stmt, child ast.Node, key string) (guarded bool, alias string, settled bool) {
	idx := -1
	for i, s := range stmts {
		if s == child {
			idx = i
			break
		}
	}
	for i := idx - 1; i >= 0; i-- {
		switch s := stmts[i].(type) {
		case *ast.IfStmt:
			if terminates(s.Body) && !assignsPath(pass, s.Else, key) {
				// code after the if only runs when the condition was false (via a
				// non-terminating else, provided it left key alone)
				for _, k := range equivalentKeys(pass, parents, s, key) {
					if impliedByNil(pass, s.Cond, k) {
						return true, "", false
					}
				}
			} else if s.Else == nil && impliedByNil(pass, s.Cond, key) && assignsNonNil(pass, s.Body.List, key) {
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
					if alias != "" && assignsPath(pass, between, alias) {
						return false, "", true
					}
				}
				return guarded, alias, settled
			}
		}
	}
	return false, "", false
}

// assignsPath reports whether any assignment or inc/dec anywhere inside n targets key or a
// prefix of it; nil n never does.
func assignsPath(pass *analysis.Pass, n ast.Node, key string) bool {
	if n == nil {
		return false
	}
	found := false
	ast.Inspect(n, func(x ast.Node) bool {
		var lhs []ast.Expr
		switch s := x.(type) {
		case *ast.AssignStmt:
			lhs = s.Lhs
		case *ast.IncDecStmt:
			lhs = []ast.Expr{s.X}
		}
		for _, e := range lhs {
			if k, ok := PathKey(pass, e); ok && (k == key || strings.HasPrefix(key, k+".")) {
				found = true
			}
		}
		return !found
	})
	return found
}

// assignsNonNil reports whether one of stmts assigns key a provably non-nil value.
func assignsNonNil(pass *analysis.Pass, stmts []ast.Stmt, key string) bool {
	for _, s := range stmts {
		assign, ok := s.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != len(assign.Rhs) {
			continue
		}
		for i, lhs := range assign.Lhs {
			if lk, ok := PathKey(pass, lhs); ok && lk == key && isNonNilSource(pass, assign.Rhs[i]) {
				return true
			}
		}
	}
	return false
}

// assignmentGuard classifies what assign does to key; matched is false when it does not touch
// key or a prefix of it. A provable source proves the guard, an alias of another chain
// redirects the search, a multi-result call is guarded when its error/ok companion was
// checked in following (the statements up to the dereference, or the if the call is the init
// of), and anything else settles the value as unknown.
func assignmentGuard(pass *analysis.Pass, assign *ast.AssignStmt, key string, following []ast.Stmt) (guarded bool, alias string, settled, matched bool) {
	for j, lhs := range assign.Lhs {
		lk, ok := PathKey(pass, lhs)
		if !ok {
			continue
		}
		if lk != key && !strings.HasPrefix(key, lk+".") {
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
					if nonNilResult(pass, calledFunc(pass, call), j) {
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

// literalField returns the value a composite literal (or &literal) gives the field path in
// suffix (".F.G"), descending nested literals; nil when expr is not a literal, a segment is
// absent (the zero value) or unkeyed, or an intermediate value is not itself a literal.
func literalField(expr ast.Expr, suffix string) ast.Expr {
	for seg := range strings.SplitSeq(strings.TrimPrefix(suffix, "."), ".") {
		e := ast.Unparen(expr)
		if u, ok := e.(*ast.UnaryExpr); ok && u.Op == token.AND {
			e = ast.Unparen(u.X)
		}
		lit, ok := e.(*ast.CompositeLit)
		if !ok {
			return nil
		}
		expr = nil
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if id, ok := kv.Key.(*ast.Ident); ok && id.Name == seg {
				expr = kv.Value
				break
			}
		}
		if expr == nil {
			return nil
		}
	}
	return expr
}

// compositeIDArg returns the argument a commonids composite-ID constructor or parser stores
// under suffix (".First" or ".Second"), nil when expr is neither call or suffix is another
// field. Both functions take the two IDs as their trailing arguments.
func compositeIDArg(pass *analysis.Pass, expr ast.Expr, suffix string) ast.Expr {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok || len(call.Args) < 2 {
		return nil
	}
	fn := calledFunc(pass, call)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != commonIDsPkgPath ||
		(fn.Name() != "NewCompositeResourceID" && fn.Name() != "ParseCompositeResourceID") {
		return nil
	}
	switch suffix {
	case ".First":
		return call.Args[len(call.Args)-2]
	case ".Second":
		return call.Args[len(call.Args)-1]
	}
	return nil
}

const commonIDsPkgPath = "github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"

// calledFunc resolves the function a call invokes, through parens and explicit generic
// instantiation; nil for builtins, function values, and method values on unresolved types.
func calledFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	fun := ast.Unparen(call.Fun)
	switch ix := fun.(type) {
	case *ast.IndexExpr:
		fun = ix.X
	case *ast.IndexListExpr:
		fun = ix.X
	}
	var id *ast.Ident
	switch f := fun.(type) {
	case *ast.Ident:
		id = f
	case *ast.SelectorExpr:
		id = f.Sel
	default:
		return nil
	}
	fn, _ := pass.TypesInfo.Uses[id].(*types.Func)
	return fn
}

// companionKeys returns the keys of the error and ok-bool results assign binds alongside its
// other results.
func companionKeys(pass *analysis.Pass, assign *ast.AssignStmt) (errKey, okKey string) {
	errType := types.Universe.Lookup("error").Type()
	for _, lhs := range assign.Lhs {
		id, ok := ast.Unparen(lhs).(*ast.Ident)
		if !ok {
			continue
		}
		obj := pass.TypesInfo.Uses[id]
		if obj == nil {
			obj = pass.TypesInfo.Defs[id]
		}
		if obj == nil {
			continue
		}
		if types.Identical(obj.Type(), errType) {
			errKey, _ = PathKey(pass, lhs)
		} else if basic, isBasic := obj.Type().Underlying().(*types.Basic); isBasic && basic.Kind() == types.Bool {
			okKey, _ = PathKey(pass, lhs)
		}
	}
	return errKey, okKey
}

// companionCheckedBetween reports whether assign returns an error or ok-bool alongside the
// pointer and one of the following statements exits on it (`if err != nil { return ... }`,
// `if !ok { return ... }`) — the conventional Go contracts under which the other results are
// valid.
func companionCheckedBetween(pass *analysis.Pass, assign *ast.AssignStmt, between []ast.Stmt) bool {
	errKey, okKey := companionKeys(pass, assign)
	for _, s := range between {
		if errKey == "" && okKey == "" {
			return false
		}
		switch s := s.(type) {
		case *ast.IfStmt:
			if s.Else == nil && terminates(s.Body) && condProvesInvalid(pass, s.Cond, errKey, okKey) {
				return true
			}
		case *ast.AssignStmt:
			// a companion overwritten before the check no longer speaks for the call
			for _, lhs := range s.Lhs {
				switch k, _ := PathKey(pass, lhs); k {
				case errKey:
					errKey = ""
				case okKey:
					okKey = ""
				}
			}
		}
	}
	return false
}

// condProvesInvalid reports whether cond being true means the companion signalled failure —
// `err != nil` or `!ok`, possibly inside a pure-|| chain — so cond being false (an exit not
// taken, an else branch) proves the other results valid.
func condProvesInvalid(pass *analysis.Pass, cond ast.Expr, errKey, okKey string) bool {
	return (errKey != "" && orContainsNeqNil(pass, cond, errKey)) || (okKey != "" && orContainsNot(pass, cond, okKey))
}

// condProvesValid reports whether cond being true means the companion signalled success —
// `err == nil` or `ok`, possibly inside a pure-&& chain.
func condProvesValid(pass *analysis.Pass, cond ast.Expr, errKey, okKey string) bool {
	if errKey == "" && okKey == "" {
		return false
	}
	switch x := ast.Unparen(cond).(type) {
	case *ast.BinaryExpr:
		if x.Op == token.LAND {
			return condProvesValid(pass, x.X, errKey, okKey) || condProvesValid(pass, x.Y, errKey, okKey)
		}
		return errKey != "" && x.Op == token.EQL && nilComparison(pass, x, errKey)
	case *ast.Ident:
		k, ok := PathKey(pass, x)
		return ok && okKey != "" && k == okKey
	}
	return false
}

// orContainsNot reports whether cond is `!ok` or a pure-|| chain containing it.
func orContainsNot(pass *analysis.Pass, cond ast.Expr, key string) bool {
	switch x := ast.Unparen(cond).(type) {
	case *ast.BinaryExpr:
		if x.Op == token.LOR {
			return orContainsNot(pass, x.X, key) || orContainsNot(pass, x.Y, key)
		}
	case *ast.UnaryExpr:
		if x.Op == token.NOT {
			k, ok := PathKey(pass, x.X)
			return ok && k == key
		}
	}
	return false
}

// orContainsNeqNil reports whether cond is `key != nil` or a pure-|| chain containing it, so
// cond being false proves key == nil.
func orContainsNeqNil(pass *analysis.Pass, cond ast.Expr, key string) bool {
	x, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if x.Op == token.LOR {
		return orContainsNeqNil(pass, x.X, key) || orContainsNeqNil(pass, x.Y, key)
	}
	return x.Op == token.NEQ && nilComparison(pass, x, key)
}

// terminates reports whether a block's final statement leaves the enclosing flow: a return,
// branch, panic, or fatal call.
func terminates(block *ast.BlockStmt) bool {
	if len(block.List) == 0 {
		return false
	}
	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	case *ast.ExprStmt:
		call, ok := last.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			return fun.Name == "panic"
		case *ast.SelectorExpr:
			name := fun.Sel.Name
			return name == "Exit" || strings.HasPrefix(name, "Fatal")
		}
	}
	return false
}

// isNonNilSource reports whether expr can never be nil: an address-of, new(...), a
// pointer.To*(...) call (always allocates), a flag package constructor (returns a pointer
// into the flag set), an immediately-invoked func literal whose returns all qualify, or a
// call to a function ReturnsAnalyzer proved never returns nil in that position.
func isNonNilSource(pass *analysis.Pass, expr ast.Expr) bool {
	switch x := ast.Unparen(expr).(type) {
	case *ast.UnaryExpr:
		return x.Op == token.AND
	case *ast.CallExpr:
		if id, ok := ast.Unparen(x.Fun).(*ast.Ident); ok {
			if _, isBuiltin := pass.TypesInfo.Uses[id].(*types.Builtin); isBuiltin {
				return id.Name == "new"
			}
		}
		// an immediately-invoked func literal whose every return is itself non-nil
		if lit, ok := ast.Unparen(x.Fun).(*ast.FuncLit); ok {
			return funcLitReturnsNonNil(pass, lit)
		}
		fn := calledFunc(pass, x)
		if fn == nil || fn.Pkg() == nil {
			return false
		}
		return (fn.Pkg().Path() == pointerpkg.PkgPath && strings.HasPrefix(fn.Name(), "To")) ||
			fn.Pkg().Path() == "flag" || nonNilResult(pass, fn, 0)
	}
	return false
}

// funcLitReturnsNonNil reports whether lit has a single result and every return statement in
// its body (nested literals excluded) returns a provably non-nil value.
func funcLitReturnsNonNil(pass *analysis.Pass, lit *ast.FuncLit) bool {
	if lit.Type.Results == nil || len(lit.Type.Results.List) != 1 || len(lit.Type.Results.List[0].Names) > 1 {
		return false
	}
	returns := 0
	nonNil := true
	ast.Inspect(lit.Body, func(n ast.Node) bool {
		switch r := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			returns++
			if len(r.Results) != 1 || !isNonNilSource(pass, r.Results[0]) {
				nonNil = false
			}
		}
		return nonNil
	})
	return returns > 0 && nonNil
}

// impliesNonNil reports whether cond being true proves key != nil.
func impliesNonNil(pass *analysis.Pass, cond ast.Expr, key string) bool {
	x, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if x.Op == token.LAND {
		return impliesNonNil(pass, x.X, key) || impliesNonNil(pass, x.Y, key)
	}
	return (x.Op == token.NEQ && nilComparison(pass, x, key)) || fromNonZero(pass, x, key)
}

// fromNonZero reports whether cmp is `pointer.From(x) != <zero>` or `len(pointer.From(x)) > 0`
// (either operand order) for key's path — both false when x is nil, so true proves x non-nil.
func fromNonZero(pass *analysis.Pass, cmp *ast.BinaryExpr, key string) bool {
	matches := func(e, other ast.Expr, op token.Token) bool {
		call, ok := ast.Unparen(e).(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return false
		}
		if fn := calledFunc(pass, call); fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == pointerpkg.PkgPath && fn.Name() == "From" {
			k, ok := PathKey(pass, call.Args[0])
			return ok && k == key && op == token.NEQ && isZeroConst(pass, other)
		}
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "len" {
			if _, isBuiltin := pass.TypesInfo.Uses[id].(*types.Builtin); !isBuiltin {
				return false
			}
			inner, ok := ast.Unparen(call.Args[0]).(*ast.CallExpr)
			if !ok || len(inner.Args) != 1 {
				return false
			}
			fn := calledFunc(pass, inner)
			if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != pointerpkg.PkgPath || fn.Name() != "From" {
				return false
			}
			k, ok := PathKey(pass, inner.Args[0])
			return ok && k == key && (op == token.NEQ || op == token.GTR) && isZeroConst(pass, other)
		}
		return false
	}
	flipped := map[token.Token]token.Token{token.NEQ: token.NEQ, token.GTR: token.LSS, token.LSS: token.GTR}
	return matches(cmp.X, cmp.Y, cmp.Op) || matches(cmp.Y, cmp.X, flipped[cmp.Op])
}

// isZeroConst reports whether e is a constant "" or 0.
func isZeroConst(pass *analysis.Pass, e ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[e]
	if !ok || tv.Value == nil {
		return false
	}
	switch tv.Value.Kind() {
	case constant.String:
		return constant.StringVal(tv.Value) == ""
	case constant.Int, constant.Float:
		return constant.Sign(tv.Value) == 0
	case constant.Unknown, constant.Bool, constant.Complex:
		return false
	}
	return false
}

// impliedByNil reports whether key == nil forces cond to be true — so cond being false (an
// else branch, a short-circuited ||, a not-taken early exit) proves key != nil.
func impliedByNil(pass *analysis.Pass, cond ast.Expr, key string) bool {
	x, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if x.Op == token.LOR {
		return impliedByNil(pass, x.X, key) || impliedByNil(pass, x.Y, key)
	}
	return x.Op == token.EQL && nilComparison(pass, x, key)
}

// nilComparison reports whether cmp compares key's path against nil (either operand order).
func nilComparison(pass *analysis.Pass, cmp *ast.BinaryExpr, key string) bool {
	matches := func(e, other ast.Expr) bool {
		k, ok := PathKey(pass, e)
		return ok && k == key && astx.IsNilValue(pass, other)
	}
	return matches(cmp.X, cmp.Y) || matches(cmp.Y, cmp.X)
}
