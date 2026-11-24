// Package description finds the descriptions written out in a provider's code, and maps a place in a description's text back to the source so a
// check can report and fix it there.
package description

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/facts"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer finds the descriptions in a package once, for every check that reads them. It also records which functions take a description as a
// parameter, so the text handed to a shared schema helper is read where it is written, in whichever package that is.
var Analyzer = &analysis.Analyzer{
	Name:       "descriptions",
	Doc:        "find the resource and property descriptions written out in a package",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	FactTypes:  []analysis.Fact{(*takesDescription)(nil)},
	ResultType: reflect.TypeFor[*Descriptions](),
}

// run refers to Analyzer, so it is wired up here and not in the literal, which would be an initialization cycle.
func init() { Analyzer.Run = run }

// takesDescription is the fact on a function that builds a schema from what it is given: bit i is set when parameter i becomes a description.
type takesDescription struct{ Params uint64 }

func (*takesDescription) AFact() {}

func (f *takesDescription) String() string { return "description(" + facts.Bits(f.Params) + ")" }

// Descriptions is what Analyzer finds in a package.
type Descriptions struct {
	// Found is every description written as text.
	Found []Found

	// Helpers holds, for each function of the package that takes a description, the schema value each such parameter is set on.
	Helpers map[*types.Func]map[int]*ast.CompositeLit
}

// All returns every description in the package, for a check that requires Analyzer.
func All(pass *analysis.Pass) []Found {
	if found, ok := pass.ResultOf[Analyzer].(*Descriptions); ok {
		return found.Found
	}

	return nil
}

// Found is one description written as text in the code: on a schema property, a resource, a framework attribute or block, a provider function or
// its parameters, or returned by a typed resource's Description() method.
type Found struct {
	// Value is the expression holding the text.
	Value ast.Expr

	// Text is what it says.
	Text string

	// Owner is the schema value the description is set on, whose other fields say what the property is. It is nil for a typed resource's method,
	// and for text handed to a helper in another package.
	Owner *ast.CompositeLit

	// Via is the helper the text is handed to, and Param which of its parameters takes it. Via is nil for a description set where it is written.
	Via   *types.Func
	Param int

	// where each byte of Text sits in the source, counted from the literal's opening quote; nil when Value is not one string literal
	offsets []int

	// raw is whether the literal is written in backticks, where nothing can be escaped
	raw bool
}

// Finding is one thing wrong with a description's text.
type Finding struct {
	// Start and End are the bytes of the text it is about; the same number twice is a place between two bytes.
	Start, End int

	Message string

	// Fix names the fix, and New is what it puts in place of the bytes. An empty Fix means there is none.
	Fix, New string
}

// run finds every description in the package, outside its test files, that is text the compiler can read: a string, or constants joined together.
// One handed to a function that sets it on a schema is found at the call.
func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return &Descriptions{}, nil
	}

	// the plugin sdk and the framework keep their schema types in a package named schema, and a provider function's in one named function; an
	// Azure API model with a Description field is neither
	described := func(lit *ast.CompositeLit) []*ast.KeyValueExpr {
		t := pass.TypesInfo.TypeOf(lit)
		if t == nil {
			return nil
		}
		if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
			t = ptr.Elem()
		}

		named, ok := types.Unalias(t).(*types.Named)
		if !ok || named.Obj().Pkg() == nil || !strings.Contains(named.Obj().Pkg().Path(), "/terraform-plugin-") {
			return nil
		}
		if name := named.Obj().Pkg().Name(); name != "schema" && name != "function" {
			return nil
		}

		var descriptions []*ast.KeyValueExpr
		for _, elt := range lit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if key, ok := kv.Key.(*ast.Ident); ok && (key.Name == "Description" || key.Name == "MarkdownDescription") {
					descriptions = append(descriptions, kv)
				}
			}
		}

		return descriptions
	}

	// a function takes a description when it sets a parameter as one, or hands it to a function that does
	type helper struct {
		fn     *types.Func
		body   *ast.BlockStmt
		params map[types.Object]int
	}
	helpers := map[*types.Func]map[int]*ast.CompositeLit{}
	takes, err := facts.Run(pass, Analyzer,
		func(fn *types.Func, decl *ast.FuncDecl) (helper, bool) {
			h := helper{fn: fn, body: decl.Body, params: map[types.Object]int{}}
			i := 0
			for _, field := range decl.Type.Params.List {
				for _, name := range field.Names {
					if obj := pass.TypesInfo.Defs[name]; obj != nil && i < 64 {
						h.params[obj] = i
					}
					i++
				}
				if len(field.Names) == 0 {
					i++
				}
			}

			return h, len(h.params) > 0 && !strings.HasSuffix(pass.Fset.Position(decl.Pos()).Filename, "_test.go")
		},
		func(s *facts.Set[takesDescription, *takesDescription], h helper) takesDescription {
			param := func(e ast.Expr) (int, bool) {
				id, ok := ast.Unparen(e).(*ast.Ident)
				if !ok {
					return 0, false
				}
				i, ok := h.params[pass.TypesInfo.Uses[id]]

				return i, ok
			}

			var f takesDescription
			ast.Inspect(h.body, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.CompositeLit:
					for _, kv := range described(n) {
						if i, ok := param(kv.Value); ok {
							f.Params |= 1 << i
							if helpers[h.fn] == nil {
								helpers[h.fn] = map[int]*ast.CompositeLit{}
							}
							helpers[h.fn][i] = n
						}
					}

				case *ast.CallExpr:
					callee := s.Get(astx.CalledFunc(pass, n))
					for j, arg := range n.Args {
						if i, ok := param(arg); ok && j < 64 && callee.Params&(1<<j) != 0 {
							f.Params |= 1 << i
						}
					}
				}

				return true
			})

			return f
		})
	if err != nil {
		return nil, err
	}

	var found []Found
	add := func(value ast.Expr, owner *ast.CompositeLit, via *types.Func, param int) {
		tv, ok := pass.TypesInfo.Types[value]
		if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
			return
		}

		f := Found{Value: value, Text: constant.StringVal(tv.Value), Owner: owner, Via: via, Param: param}
		lit, ok := ast.Unparen(value).(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || len(lit.Value) < 2 {
			found = append(found, f)
			return
		}

		f.Value = lit
		f.raw = lit.Value[0] == '`'
		f.offsets = make([]int, 0, len(f.Text)+1)
		src := lit.Value[1 : len(lit.Value)-1]
		for at := 0; at < len(src); {
			if f.raw {
				f.offsets = append(f.offsets, at+1)
				at++
				continue
			}

			r, multibyte, tail, err := strconv.UnquoteChar(src[at:], '"')
			if err != nil {
				f.offsets = nil
				break
			}

			// an escape is one place in the source however many bytes it stands for; a character typed as itself is as wide as it is
			used, width := len(src)-at-len(tail), 1
			if multibyte {
				width = utf8.RuneLen(r)
			}
			for i := range width {
				if used == width {
					f.offsets = append(f.offsets, at+1+i)
				} else {
					f.offsets = append(f.offsets, at+1)
				}
			}
			at += used
		}

		if f.offsets != nil {
			f.offsets = append(f.offsets, len(src)+1)
		}
		if len(f.offsets) != len(f.Text)+1 {
			f.offsets = nil
		}
		found = append(found, f)
	}

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil), (*ast.FuncDecl)(nil), (*ast.CallExpr)(nil)}, func(n ast.Node) {
		if strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return
		}

		switch n := n.(type) {
		case *ast.CompositeLit:
			for _, kv := range described(n) {
				add(kv.Value, n, nil, 0)
			}

		case *ast.CallExpr:
			fn := astx.CalledFunc(pass, n)
			for j, arg := range n.Args {
				if j < 64 && takes.Get(fn).Params&(1<<j) != 0 {
					add(arg, helpers[fn][j], fn, j)
				}
			}

		case *ast.FuncDecl:
			if n.Recv == nil || n.Name.Name != "Description" || n.Body == nil || len(n.Body.List) != 1 {
				return
			}

			fn, ok := pass.TypesInfo.Defs[n.Name].(*types.Func)
			if !ok {
				return
			}

			sig := fn.Signature()
			if sig.Params().Len() != 0 || sig.Results().Len() != 1 || !types.Identical(sig.Results().At(0).Type(), types.Typ[types.String]) {
				return
			}

			// every type the sdk registers says what it is called, which tells its Description() from any other
			if obj, _, _ := types.LookupFieldOrMethod(sig.Recv().Type(), true, pass.Pkg, "ResourceType"); obj == nil {
				return
			}

			if ret, ok := n.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
				add(ret.Results[0], nil, nil, 0)
			}
		}
	})

	return &Descriptions{Found: found, Helpers: helpers}, nil
}

// Report reports a finding at its place inside the description, with its fix. A description that is not one string literal is reported as a whole,
// without a fix, and so is one whose fix would rewrite what another rule's fix already does.
func (f Found) Report(pass *analysis.Pass, finding Finding) {
	if f.offsets == nil {
		pass.Report(analysis.Diagnostic{Pos: f.Value.Pos(), End: f.Value.End(), Message: finding.Message})
		return
	}

	d := analysis.Diagnostic{Pos: f.Value.Pos() + token.Pos(f.offsets[finding.Start]), End: f.Value.Pos() + token.Pos(f.offsets[finding.End]), Message: finding.Message}
	text := finding.New
	if !f.raw {
		text = strconv.Quote(text)
		text = text[1 : len(text)-1]
	}

	if finding.Fix != "" && (!f.raw || !strings.Contains(text, "`")) && astx.Claim(pass, d.Pos, d.End) {
		d.SuggestedFixes = []analysis.SuggestedFix{{Message: finding.Fix, TextEdits: []analysis.TextEdit{{Pos: d.Pos, End: d.End, NewText: []byte(text)}}}}
	}
	pass.Report(d)
}
