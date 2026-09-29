package nilguard

import (
	"go/ast"
	"go/types"
	"reflect"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/facts"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// ReturnsAnalyzer records, for every function in a package, which pointer results are never
// nil — every return statement supplies a provably non-nil value in that position: an
// address-of, new(T), pointer.To*(...), a call to a function already known to return non-nil
// there, or a local the guard engine proves non-nil at the return. The verdicts are exported
// as facts so callers in other packages can treat such calls as non-nil sources. Guard checks
// consult it through the pass, so any analyzer using the engine must list it in Requires.
var ReturnsAnalyzer = &analysis.Analyzer{
	Name:       "nilguard_returns",
	Doc:        "record which pointer results of each function are never nil",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	FactTypes:  []analysis.Fact{(*nonNilResults)(nil)},
	ResultType: reflect.TypeFor[*returnsSet](),
}

// runReturns refers to ReturnsAnalyzer, so it is wired up here rather than in the literal to
// avoid an initialization cycle.
func init() { ReturnsAnalyzer.Run = runReturns }

// nonNilResults is the fact attached to a function: bit i set means result i is never nil.
type nonNilResults struct{ Mask uint64 }

type returnsSet = facts.Set[nonNilResults, *nonNilResults]

func (*nonNilResults) AFact() {}

func (f *nonNilResults) String() string { return "nonnil(" + facts.Bits(f.Mask) + ")" }

// nonNilResult is the engine's hook: true when the returns verdicts are available to pass and
// vouch for fn's i-th result.
func nonNilResult(pass *analysis.Pass, fn *types.Func, i int) bool {
	return i >= 0 && i < 64 && facts.Of[nonNilResults](pass, ReturnsAnalyzer).Get(fn).Mask&(1<<i) != 0
}

// returnsCandidate is what a declaration contributes to the fixpoint: its body's parent map,
// its named result idents by position (nil when unnamed), and the pointer-typed positions.
type returnsCandidate struct {
	body    *ast.BlockStmt
	parents map[ast.Node]ast.Node
	results []*ast.Ident
	mask    uint64
}

func runReturns(pass *analysis.Pass) (any, error) {
	return facts.Run(pass, ReturnsAnalyzer,
		func(_ *types.Func, decl *ast.FuncDecl) (returnsCandidate, bool) {
			c := returnsCandidate{body: decl.Body}
			if decl.Type.Results == nil {
				return c, false
			}
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
			if c.mask == 0 {
				return c, false
			}
			c.parents = ParentMap(decl.Body)
			return c, true
		},
		func(_ *returnsSet, c returnsCandidate) nonNilResults {
			mask := c.mask
			ast.Inspect(c.body, func(n ast.Node) bool {
				switch ret := n.(type) {
				case *ast.FuncLit:
					return false
				case *ast.ReturnStmt:
					mask &= returnMask(pass, c.parents, ret, c.results)
				}
				return mask != 0
			})
			return nonNilResults{Mask: mask}
		})
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

// ParentMap returns the child-to-parent map for every node under root.
func ParentMap(root ast.Node) map[ast.Node]ast.Node {
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
