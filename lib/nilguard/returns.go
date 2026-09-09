package nilguard

import (
	"go/ast"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ReturnsAnalyzer records, for every function in a package, which pointer results are never
// nil — every return statement supplies a provably non-nil value in that position: an
// address-of, new(T), pointer.To*(...), a call to a function already known to return non-nil
// there, or a local the guard engine proves non-nil at the return. The verdicts are exported
// as facts so callers in other packages can treat such calls as non-nil sources, and returned
// as a Returns for the current package. Guard checks consult it through the pass, so any
// analyzer using the engine must list it in Requires.
var ReturnsAnalyzer = &analysis.Analyzer{
	Name:       "nilguard_returns",
	Doc:        "record which pointer results of each function are never nil",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	FactTypes:  []analysis.Fact{(*nonNilResults)(nil)},
	ResultType: reflect.TypeFor[*Returns](),
}

// runReturns refers to ReturnsAnalyzer (to publish its in-progress result), so it is wired
// up here rather than in the literal to avoid an initialization cycle.
func init() { ReturnsAnalyzer.Run = runReturns }

// nonNilResults is the fact attached to a function: bit i set means result i is never nil.
type nonNilResults struct{ Mask uint64 }

func (*nonNilResults) AFact() {}

func (f *nonNilResults) String() string {
	var idx []string
	for i := range 64 {
		if f.Mask&(1<<i) != 0 {
			idx = append(idx, strconv.Itoa(i))
		}
	}
	return "nonnil(" + strings.Join(idx, ",") + ")"
}

// Returns answers non-nil-result queries for the package being analysed and, via facts, for
// its dependencies.
type Returns struct {
	pass  *analysis.Pass
	local map[*types.Func]uint64
}

// NonNil reports whether fn's i-th result is never nil.
func (r *Returns) NonNil(fn *types.Func, i int) bool {
	if fn == nil || i < 0 || i > 63 {
		return false
	}
	fn = fn.Origin()
	if fn.Pkg() == r.pass.Pkg {
		return r.local[fn]&(1<<i) != 0
	}
	var fact nonNilResults
	return r.pass.ImportObjectFact(fn, &fact) && fact.Mask&(1<<i) != 0
}

// nonNilResult is the engine's hook: true when the pass has the returns result available and
// it vouches for fn's i-th result.
func nonNilResult(pass *analysis.Pass, fn *types.Func, i int) bool {
	r, ok := pass.ResultOf[ReturnsAnalyzer].(*Returns)
	return ok && r.NonNil(fn, i)
}

func runReturns(pass *analysis.Pass) (any, error) {
	r := &Returns{pass: pass, local: map[*types.Func]uint64{}}
	// make the in-progress verdicts visible to the guard engine during the fixpoint below,
	// the same way dependents see the finished result
	pass.ResultOf[ReturnsAnalyzer] = r

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return r, nil
	}

	type candidate struct {
		fn      *types.Func
		decl    *ast.FuncDecl
		parents map[ast.Node]ast.Node
		results []*ast.Ident // named result idents by position, nil entries when unnamed
		mask    uint64       // pointer-typed result positions
	}
	var candidates []*candidate
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, ok := n.(*ast.FuncDecl)
		if !ok || decl.Body == nil || decl.Type.Results == nil {
			return
		}
		fn, _ := pass.TypesInfo.Defs[decl.Name].(*types.Func)
		if fn == nil {
			return
		}
		c := &candidate{fn: fn, decl: decl}
		for _, field := range decl.Type.Results.List {
			names := field.Names
			if len(names) == 0 {
				names = []*ast.Ident{nil}
			}
			for _, name := range names {
				i := len(c.results)
				c.results = append(c.results, name)
				if t := pass.TypesInfo.TypeOf(field.Type); t != nil && i < 64 {
					if _, isPtr := t.Underlying().(*types.Pointer); isPtr {
						c.mask |= 1 << i
					}
				}
			}
		}
		if c.mask != 0 {
			c.parents = parentMap(decl.Body)
			candidates = append(candidates, c)
		}
	})

	// verdicts only ever grow (a newly proven callee can prove more callers), so iterate to a
	// fixpoint; it converges in a few rounds
	for changed := true; changed; {
		changed = false
		for _, c := range candidates {
			mask := c.mask
			ast.Inspect(c.decl.Body, func(n ast.Node) bool {
				switch ret := n.(type) {
				case *ast.FuncLit:
					return false
				case *ast.ReturnStmt:
					mask &= returnMask(pass, c.parents, ret, c.results)
				}
				return mask != 0
			})
			if mask != r.local[c.fn] {
				r.local[c.fn] = mask
				changed = true
			}
		}
	}

	for fn, mask := range r.local {
		if mask != 0 {
			pass.ExportObjectFact(fn, &nonNilResults{Mask: mask})
		}
	}
	return r, nil
}

// returnMask reports which positions ret provably returns non-nil: a naked return is judged
// by the named results' state at that point, `return f()` by f's verdict, and an explicit
// list by each expression — a non-nil source, or a path the engine proves non-nil there.
func returnMask(pass *analysis.Pass, parents map[ast.Node]ast.Node, ret *ast.ReturnStmt, named []*ast.Ident) uint64 {
	var mask uint64
	proven := func(i int, expr ast.Expr) {
		if isNonNilSource(pass, expr) {
			mask |= 1 << i
			return
		}
		if key, ok := PathKey(pass, expr); ok && Guarded(pass, parents, ret, key) {
			mask |= 1 << i
		}
	}
	switch {
	case len(ret.Results) == 0:
		for i, name := range named {
			if name != nil {
				proven(i, name)
			}
		}
	case len(ret.Results) == 1 && len(named) > 1:
		if call, ok := ast.Unparen(ret.Results[0]).(*ast.CallExpr); ok {
			if fn := astx.CalledFunc(pass, call); fn != nil {
				for i := range named {
					if nonNilResult(pass, fn, i) {
						mask |= 1 << i
					}
				}
			}
		}
	default:
		for i, expr := range ret.Results {
			proven(i, expr)
		}
	}
	return mask
}

// parentMap returns the child-to-parent map for every node under root.
func parentMap(root ast.Node) map[ast.Node]ast.Node {
	parents := map[ast.Node]ast.Node{}
	var stack []ast.Node
	ast.Inspect(root, func(x ast.Node) bool {
		if x == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[x] = stack[len(stack)-1]
		}
		stack = append(stack, x)
		return true
	})
	return parents
}
