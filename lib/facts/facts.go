// Package facts is the plumbing shared by analyzers that attach a fact to every function and
// propagate it across packages: the fixpoint over a package's function declarations, the
// local-or-imported lookup, and the export. What a fact means stays with the analyzer that
// computes it.
package facts

import (
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Set holds one analyzer's per-function facts: computed for the package being analysed,
// imported for its dependencies. T is the fact value, whose zero means "nothing known"; *T
// must implement analysis.Fact.
type Set[T comparable, PT interface {
	*T
	analysis.Fact
}] struct {
	pass  *analysis.Pass
	local map[*types.Func]T
}

// Get returns fn's fact — the current verdict for a function of this package (final once Run
// has returned, in progress while it iterates), the exported fact for one from a dependency —
// or the zero T. A nil Set or fn yields the zero T too.
func (s *Set[T, PT]) Get(fn *types.Func) T {
	var v T
	if s == nil || fn == nil {
		return v
	}
	fn = fn.Origin() // instantiated generics resolve to their declaration
	if fn.Pkg() == s.pass.Pkg {
		return s.local[fn]
	}
	s.pass.ImportObjectFact(fn, PT(&v))
	return v
}

// Of returns the Set analyzer produced for pass, nil when analyzer is not among the pass's
// requirements (or has not started running).
func Of[T comparable, PT interface {
	*T
	analysis.Fact
}](pass *analysis.Pass, analyzer *analysis.Analyzer) *Set[T, PT] {
	s, _ := pass.ResultOf[analyzer].(*Set[T, PT])
	return s
}

// Run computes analyzer's fact for every function declared in pass. prepare sees each
// declaration once and returns what compute needs (false skips the function); compute derives
// the fact from that and from the facts known so far, and is re-run until nothing changes, so
// it must be monotone: verdicts only ever grow. The Set is published in pass.ResultOf before
// the fixpoint, so code reached from compute can consult in-progress verdicts through Of, and
// every non-zero fact is exported once the fixpoint settles. Run is the analyzer's Run.
func Run[T comparable, PT interface {
	*T
	analysis.Fact
}, C any](pass *analysis.Pass, analyzer *analysis.Analyzer, prepare func(fn *types.Func, decl *ast.FuncDecl) (C, bool), compute func(s *Set[T, PT], c C) T) (*Set[T, PT], error) {
	s := &Set[T, PT]{pass: pass, local: map[*types.Func]T{}}
	pass.ResultOf[analyzer] = s

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return s, nil
	}
	type candidate struct {
		fn *types.Func
		c  C
	}
	var candidates []candidate
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, ok := n.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return
		}
		fn, _ := pass.TypesInfo.Defs[decl.Name].(*types.Func)
		if fn == nil {
			return
		}
		if c, ok := prepare(fn, decl); ok {
			candidates = append(candidates, candidate{fn: fn, c: c})
		}
	})

	for changed := true; changed; {
		changed = false
		for _, c := range candidates {
			if v := compute(s, c.c); v != s.local[c.fn] {
				s.local[c.fn] = v
				changed = true
			}
		}
	}

	var zero T
	for fn, v := range s.local {
		if v != zero {
			pass.ExportObjectFact(fn, PT(&v))
		}
	}
	return s, nil
}

// Bits renders the set bits of mask as "0,2,5" — the conventional fact string for a mask over
// parameter or result positions.
func Bits(mask uint64) string {
	var idx []string
	for i := range 64 {
		if mask&(1<<i) != 0 {
			idx = append(idx, strconv.Itoa(i))
		}
	}
	return strings.Join(idx, ",")
}
