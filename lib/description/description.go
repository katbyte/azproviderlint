// Package description finds the descriptions written out in a provider's code, and maps a place in a description's text back to the source so a
// check can report and fix it there.
package description

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// Found is one description written as text in the code: on a schema property, a resource, a framework attribute or block, a provider function or
// its parameters, or returned by a typed resource's Description() method.
type Found struct {
	// Value is the expression holding the text.
	Value ast.Expr

	// Text is what it says.
	Text string

	// Owner is the schema value the description is set on, whose other fields say what the property is. It is nil for a typed resource's method.
	Owner *ast.CompositeLit

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

// All returns every description in the package, outside its test files, that is text the compiler can read: a string, or constants joined together.
func All(pass *analysis.Pass, insp *inspector.Inspector) []Found {
	var found []Found
	add := func(value ast.Expr, owner *ast.CompositeLit) {
		tv, ok := pass.TypesInfo.Types[value]
		if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
			return
		}

		f := Found{Value: value, Text: constant.StringVal(tv.Value), Owner: owner}
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

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil), (*ast.FuncDecl)(nil)}, func(n ast.Node) {
		if strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return
		}

		switch n := n.(type) {
		case *ast.CompositeLit:
			t := pass.TypesInfo.TypeOf(n)
			if t == nil {
				return
			}
			if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
				t = ptr.Elem()
			}

			// the plugin sdk and the framework keep their schema types in a package named schema, and a provider function's in one named
			// function; an Azure API model with a Description field is neither
			named, ok := types.Unalias(t).(*types.Named)
			if !ok || named.Obj().Pkg() == nil || !strings.Contains(named.Obj().Pkg().Path(), "/terraform-plugin-") {
				return
			}
			if name := named.Obj().Pkg().Name(); name != "schema" && name != "function" {
				return
			}

			for _, elt := range n.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				if key, ok := kv.Key.(*ast.Ident); ok && (key.Name == "Description" || key.Name == "MarkdownDescription") {
					add(kv.Value, n)
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
				add(ret.Results[0], nil)
			}
		}
	})

	return found
}

// Report reports a finding at its place inside the description, with its fix. A description that is not one string literal is reported as a whole,
// without a fix.
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

	if finding.Fix != "" && (!f.raw || !strings.Contains(text, "`")) {
		d.SuggestedFixes = []analysis.SuggestedFix{{Message: finding.Fix, TextEdits: []analysis.TextEdit{{Pos: d.Pos, End: d.End, NewText: []byte(text)}}}}
	}
	pass.Report(d)
}
