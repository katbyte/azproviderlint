// Package requestbody records which parameters of a function are sent as the body of an HTTP
// write (PUT, PATCH, POST), so callers can tell when a value they pass becomes a request
// payload.
package requestbody

import (
	"go/ast"
	"go/token"
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
				case *ast.BasicLit:
					if x.Kind == token.STRING {
						switch x.Value {
						case `"PUT"`, `"PATCH"`, `"POST"`:
							f.Writes = true
						}
					}
				case *ast.Ident:
					if isHTTPWriteMethod(pass, x) {
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

// Sinks returns the variables whose value reaches a write-body argument anywhere in body —
// passed whole, by address, or through a field — so a value copied into one of them, or into
// one of its fields, is on its way into a request.
func Sinks(pass *analysis.Pass, body ast.Node) map[types.Object]bool {
	sinks := map[types.Object]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			for j, a := range call.Args {
				if CallBodyArg(pass, call, j) {
					if o := rootObj(pass, a); o != nil {
						sinks[o] = true
					}
				}
			}
		}
		return true
	})
	return sinks
}

// SentAsBody reports whether expr's value becomes a request body: climbing out through
// parens, calls, composite literals, and address-of, it lands on a write-body argument or on
// an assignment into a sink — whole, or one of its fields. Any call the value passes through
// (a conversion, pointer.To, expandFoo) is assumed to carry it into its result; the
// conservative reading only ever withholds a fix.
func SentAsBody(pass *analysis.Pass, parents map[ast.Node]ast.Node, sinks map[types.Object]bool, expr ast.Expr) bool {
	node := ast.Node(expr)
	for {
		switch p := parents[node].(type) {
		case *ast.ParenExpr, *ast.CompositeLit, *ast.KeyValueExpr:
			node = p
		case *ast.UnaryExpr:
			if p.Op != token.AND {
				return false
			}
			node = p
		case *ast.CallExpr:
			isArg := false
			for i, a := range p.Args {
				if a == node {
					isArg = true
					if CallBodyArg(pass, p, i) {
						return true
					}
				}
			}
			if !isArg {
				return false
			}
			node = p
		case *ast.AssignStmt:
			if len(p.Lhs) == len(p.Rhs) {
				for i, r := range p.Rhs {
					if r == node && sinks[rootObj(pass, p.Lhs[i])] {
						return true
					}
				}
			}
			return false
		default:
			return false
		}
	}
}

// rootObj returns the variable at the root of a chain like `m.F.G`, `&m`, `*m`, or `m[i]`;
// nil for anything else.
func rootObj(pass *analysis.Pass, e ast.Expr) types.Object {
	for {
		switch x := ast.Unparen(e).(type) {
		case *ast.Ident:
			if o := pass.TypesInfo.Uses[x]; o != nil {
				return o
			}
			return pass.TypesInfo.Defs[x]
		case *ast.SelectorExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		case *ast.UnaryExpr:
			if x.Op != token.AND {
				return nil
			}
			e = x.X
		default:
			return nil
		}
	}
}

// isWriteMethodValue reports whether e is net/http's MethodPut/MethodPatch/MethodPost or the
// equivalent string literal.
func isWriteMethodValue(pass *analysis.Pass, e ast.Expr) bool {
	switch x := ast.Unparen(e).(type) {
	case *ast.BasicLit:
		return x.Kind == token.STRING && (x.Value == `"PUT"` || x.Value == `"PATCH"` || x.Value == `"POST"`)
	case *ast.Ident:
		return isHTTPWriteMethod(pass, x)
	case *ast.SelectorExpr:
		return isHTTPWriteMethod(pass, x.Sel)
	}
	return false
}

// isHTTPWriteMethod reports whether id names net/http's MethodPut, MethodPatch, or MethodPost.
func isHTTPWriteMethod(pass *analysis.Pass, id *ast.Ident) bool {
	c, ok := pass.TypesInfo.Uses[id].(*types.Const)
	if !ok || c.Pkg() == nil || c.Pkg().Path() != "net/http" {
		return false
	}
	switch c.Name() {
	case "MethodPut", "MethodPatch", "MethodPost":
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
