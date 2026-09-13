// Package requestbody records which parameters of a function are sent as the body of an HTTP
// write (PUT, PATCH, POST), so callers can tell when a value they pass becomes a request
// payload.
package requestbody

import (
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/facts"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Analyzer exports, per function, two derived facts and combines them: whether the function
// performs an HTTP write — it references net/http's MethodPut/MethodPatch/MethodPost or a
// "PUT"/"PATCH"/"POST" literal, or calls a function that does — and which of its parameters
// are serialised — passed, directly or inside a closure, to encoding/json or encoding/xml
// marshalling, or to a function that serialises that parameter. A parameter sent as a write
// body is one that is serialised by a function that writes. Nothing here depends on SDK
// naming: go-azure-sdk's `req.Marshal(input)` and autorest's `WithJSON(v)` qualify because
// their bodies marshal, and `CreateOrUpdateThenPoll` qualifies because it delegates.
var Analyzer = &analysis.Analyzer{
	Name:       "requestbody",
	Doc:        "record which parameters of each function are sent as an HTTP write body",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	FactTypes:  []analysis.Fact{(*bodyFact)(nil)},
	ResultType: reflect.TypeFor[*bodySet](),
}

// run refers to Analyzer, so it is wired up here rather than in the literal to avoid an
// initialization cycle.
func init() { Analyzer.Run = run }

// bodyFact is the fact attached to a function: Writes when it performs (or delegates to) an
// HTTP write, Serialised bit i when parameter i (receiver excluded) is marshalled.
type bodyFact struct {
	Writes     bool
	Serialised uint64
}

type bodySet = facts.Set[bodyFact, *bodyFact]

func (*bodyFact) AFact() {}

func (f *bodyFact) String() string {
	switch {
	case f.Writes && f.Serialised != 0:
		return "body(" + facts.Bits(f.Serialised) + ")"
	case f.Writes:
		return "writes"
	}
	return "serialises(" + facts.Bits(f.Serialised) + ")"
}

// CallBodyArg reports whether call sends its arg-th argument as a write body: the callee both
// writes and serialises that parameter, or serialises it and is handed the write method at
// this call site (`do(ctx, http.MethodPut, body)`).
func CallBodyArg(pass *analysis.Pass, call *ast.CallExpr, arg int) bool {
	if arg < 0 || arg > 63 {
		return false
	}
	f := facts.Of[bodyFact](pass, Analyzer).Get(astx.CalledFunc(pass, call))
	if f.Serialised&(1<<arg) == 0 {
		return false
	}
	if f.Writes {
		return true
	}
	for _, a := range call.Args {
		if isWriteMethodValue(pass, a) {
			return true
		}
	}
	return false
}

// bodyCandidate is what a declaration contributes to the fixpoint: its body and its
// parameters by position (receiver excluded).
type bodyCandidate struct {
	body   *ast.BlockStmt
	params map[types.Object]int
}

func run(pass *analysis.Pass) (any, error) {
	return facts.Run(pass, Analyzer,
		func(_ *types.Func, decl *ast.FuncDecl) (bodyCandidate, bool) {
			c := bodyCandidate{body: decl.Body, params: map[types.Object]int{}}
			i := 0
			for _, field := range decl.Type.Params.List {
				names := field.Names
				if len(names) == 0 {
					names = []*ast.Ident{nil}
				}
				for _, name := range names {
					if name != nil && i < 64 {
						if obj := pass.TypesInfo.Defs[name]; obj != nil {
							c.params[obj] = i
						}
					}
					i++
				}
			}
			return c, true
		},
		func(s *bodySet, c bodyCandidate) bodyFact {
			// paramIndex resolves an argument — `x`, `&x`, `*x`, parenthesised — to the
			// parameter it names, or -1
			paramIndex := func(e ast.Expr) int {
				e = ast.Unparen(e)
				switch x := e.(type) {
				case *ast.UnaryExpr:
					e = ast.Unparen(x.X)
				case *ast.StarExpr:
					e = ast.Unparen(x.X)
				}
				if id, ok := e.(*ast.Ident); ok {
					if i, ok := c.params[pass.TypesInfo.Uses[id]]; ok {
						return i
					}
				}
				return -1
			}
			var f bodyFact
			// closures are walked too: a parameter captured and marshalled inside one
			// (autorest's WithJSON) is still this function's parameter being serialised
			ast.Inspect(c.body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.BasicLit, *ast.Ident, *ast.SelectorExpr:
					if e, ok := n.(ast.Expr); ok && isWriteMethodConst(pass, e) {
						f.Writes = true
					}
				case *ast.CallExpr:
					fn := astx.CalledFunc(pass, x)
					if fn == nil {
						return true
					}
					callee := s.Get(fn)
					if callee.Writes {
						f.Writes = true
					}
					for j, a := range x.Args {
						if i := paramIndex(a); i >= 0 && (isMarshaller(fn, j) || (j < 64 && callee.Serialised&(1<<j) != 0)) {
							f.Serialised |= 1 << i
						}
					}
				}
				return true
			})
			return f
		})
}

// isWriteMethodValue reports whether e is a PUT, PATCH, or POST method value: a string
// constant with that value (net/http's MethodPut/MethodPatch/MethodPost or a literal), or a
// local variable assigned one in the enclosing declaration.
func isWriteMethodValue(pass *analysis.Pass, e ast.Expr) bool {
	e = ast.Unparen(e)
	if isWriteMethodConst(pass, e) {
		return true
	}
	id, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	v, ok := pass.TypesInfo.Uses[id].(*types.Var)
	if !ok || v.Pkg() != pass.Pkg {
		return false
	}
	var decl ast.Node
	for _, file := range pass.Files {
		for _, d := range file.Decls {
			if d.Pos() <= v.Pos() && v.Pos() < d.End() {
				decl = d
			}
		}
	}
	if decl == nil {
		return false
	}
	found := false
	ast.Inspect(decl, func(n ast.Node) bool {
		var names []ast.Expr
		var values []ast.Expr
		switch x := n.(type) {
		case *ast.AssignStmt:
			names, values = x.Lhs, x.Rhs
		case *ast.ValueSpec:
			for _, id := range x.Names {
				names = append(names, id)
			}
			values = x.Values
		default:
			return !found
		}
		if len(names) != len(values) {
			return !found
		}
		for i, name := range names {
			if id, ok := name.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == v && isWriteMethodConst(pass, values[i]) {
				found = true
			}
		}
		return !found
	})
	return found
}

// isWriteMethodConst reports whether e is a string constant equal to PUT, PATCH, or POST.
func isWriteMethodConst(pass *analysis.Pass, e ast.Expr) bool {
	tv := pass.TypesInfo.Types[ast.Unparen(e)]
	if tv.Value == nil || tv.Value.Kind() != constant.String {
		return false
	}
	switch constant.StringVal(tv.Value) {
	case "PUT", "PATCH", "POST":
		return true
	}
	return false
}

// isMarshaller reports whether fn is a standard-library serialiser whose arg-th argument is
// the value being encoded: json/xml Marshal and MarshalIndent, and Encoder.Encode.
func isMarshaller(fn *types.Func, arg int) bool {
	if arg != 0 || fn.Pkg() == nil {
		return false
	}
	switch fn.Pkg().Path() {
	case "encoding/json", "encoding/xml":
		switch fn.Name() {
		case "Marshal", "MarshalIndent", "Encode":
			return true
		}
	}
	return false
}
