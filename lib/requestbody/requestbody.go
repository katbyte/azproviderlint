// Package requestbody records which parameters of a function are sent as the body of an HTTP
// write (PUT, PATCH, POST), so callers can tell when a value they pass becomes a request
// payload.
package requestbody

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
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
	ResultType: reflect.TypeFor[*Result](),
}

// run refers to Analyzer (to publish its in-progress result), so it is wired up here rather
// than in the literal to avoid an initialization cycle.
func init() { Analyzer.Run = run }

// bodyFact is the fact attached to a function: Writes when it performs (or delegates to) an
// HTTP write, Serialised bit i when parameter i (receiver excluded) is marshalled.
type bodyFact struct {
	Writes     bool
	Serialised uint64
}

func (*bodyFact) AFact() {}

func (f *bodyFact) String() string {
	var idx []string
	for i := range 64 {
		if f.Serialised&(1<<i) != 0 {
			idx = append(idx, strconv.Itoa(i))
		}
	}
	switch {
	case f.Writes && f.Serialised != 0:
		return "body(" + strings.Join(idx, ",") + ")"
	case f.Writes:
		return "writes"
	}
	return "serialises(" + strings.Join(idx, ",") + ")"
}

// Result answers write-body queries for the package being analysed and, via facts, for its
// dependencies.
type Result struct {
	pass  *analysis.Pass
	local map[*types.Func]bodyFact
}

func (r *Result) fact(fn *types.Func) bodyFact {
	if fn == nil {
		return bodyFact{}
	}
	fn = fn.Origin()
	if fn.Pkg() == r.pass.Pkg {
		return r.local[fn]
	}
	var fact bodyFact
	r.pass.ImportObjectFact(fn, &fact)
	return fact
}

// Param reports whether fn sends its i-th parameter as a write body.
func (r *Result) Param(fn *types.Func, i int) bool {
	if i < 0 || i > 63 {
		return false
	}
	f := r.fact(fn)
	return f.Writes && f.Serialised&(1<<i) != 0
}

// CallBodyArg reports whether call sends its arg-th argument as a write body: the callee both
// writes and serialises that parameter, or serialises it and is handed the write method at
// this call site (`do(ctx, http.MethodPut, body)`).
func CallBodyArg(pass *analysis.Pass, call *ast.CallExpr, arg int) bool {
	r, ok := pass.ResultOf[Analyzer].(*Result)
	if !ok || arg < 0 || arg > 63 {
		return false
	}
	f := r.fact(astx.CalledFunc(pass, call))
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

func run(pass *analysis.Pass) (any, error) {
	r := &Result{pass: pass, local: map[*types.Func]bodyFact{}}
	pass.ResultOf[Analyzer] = r

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return r, nil
	}

	type candidate struct {
		fn     *types.Func
		decl   *ast.FuncDecl
		params map[types.Object]int
	}
	var candidates []*candidate
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl, ok := n.(*ast.FuncDecl)
		if !ok || decl.Body == nil {
			return
		}
		fn, _ := pass.TypesInfo.Defs[decl.Name].(*types.Func)
		if fn == nil {
			return
		}
		c := &candidate{fn: fn, decl: decl, params: map[types.Object]int{}}
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
		candidates = append(candidates, c)
	})

	// paramIndex resolves an argument expression — `x`, `&x`, `*x`, parenthesised — to the
	// parameter it names, or -1
	paramIndex := func(c *candidate, e ast.Expr) int {
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

	// verdicts only grow (a newly proven callee proves its callers), so iterate to a fixpoint
	for changed := true; changed; {
		changed = false
		for _, c := range candidates {
			var f bodyFact
			// closures are walked too: a parameter captured and marshalled inside one
			// (autorest's WithJSON) is still this function's parameter being serialised
			ast.Inspect(c.decl.Body, func(n ast.Node) bool {
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
					callee := r.fact(fn)
					if callee.Writes {
						f.Writes = true
					}
					for j, a := range x.Args {
						if i := paramIndex(c, a); i >= 0 && (isMarshaller(fn, j) || (j < 64 && callee.Serialised&(1<<j) != 0)) {
							f.Serialised |= 1 << i
						}
					}
				}
				return true
			})
			if f != r.local[c.fn] {
				r.local[c.fn] = f
				changed = true
			}
		}
	}

	for fn, f := range r.local {
		if f.Writes || f.Serialised != 0 {
			pass.ExportObjectFact(fn, &f)
		}
	}
	return r, nil
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
