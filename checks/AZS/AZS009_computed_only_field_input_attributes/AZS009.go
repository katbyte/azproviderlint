// Package AZS009 defines an analyzer that reports computed-only schema fields setting
// attributes that only apply to user input, and Optional/Required fields nested inside a
// computed-only block.
package AZS009

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"slices"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks schema fields that are computed-only (Computed without Optional or Required,
// anything nested in such a block, and everything in a typed resource's Attributes()) for
// attributes that describe user input: validation, defaults, item counts, conflicts, and the
// Optional/Required flags themselves. The plugin SDK rejects most of these at provider start,
// but it does not see fields nested inside a computed-only block, or an element schema's
// ValidateFunc, and a lint catches all of them at the desk.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS009",
	Doc:      "check for computed-only schema fields that set input-only attributes (ValidateFunc, MaxItems, Default, ...) or nest Optional/Required fields",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS009_computed_only_field_input_attributes/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// inputOnly lists the Schema fields that only mean something for a value the user writes: how it
// is validated, defaulted, normalised, counted, or constrained against other arguments. It is the
// plugin SDK's own computed-only rejection list plus RequiredWith and WriteOnly.
var inputOnly = []string{
	"Default", "DefaultFunc", "InputDefault",
	"ValidateFunc", "ValidateDiagFunc",
	"DiffSuppressFunc", "DiffSuppressOnRefresh", "StateFunc",
	"MaxItems", "MinItems",
	"AtLeastOneOf", "ExactlyOneOf", "ConflictsWith", "RequiredWith",
	"WriteOnly",
}

const (
	onField      = "on a computed-only field"
	inBlock      = "inside a computed-only block"
	onElem       = "on the element of a computed-only field"
	inAttributes = "in Attributes(), where every field is computed"
)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// nested literals are handled from the block that makes them computed-only, so the
	// walk must not visit them again on their own
	visited := map[*ast.CompositeLit]bool{}

	insp.WithStack([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return false
		}
		cl, ok := n.(*ast.CompositeLit)
		if !ok || visited[cl] || !tf.IsSchemaHelperType(pass, cl, "Schema") {
			return true
		}
		kvs := keyValues(cl)
		switch {
		case astx.IsTrueConstant(pass, value(kvs, "Computed")) &&
			!astx.IsTrueConstant(pass, value(kvs, "Optional")) &&
			!astx.IsTrueConstant(pass, value(kvs, "Required")):
			check(pass, cl, visited, onField, false)
		case inAttributesMethod(pass, stack):
			check(pass, cl, visited, inAttributes, true)
		}
		return true
	})

	return nil, nil
}

// check reports what cl sets that a computed-only field has no use for, then descends into
// its Elem: everything under a computed-only block is computed-only too. flags also reports
// Optional/Required on cl itself, which is only wrong where cl is not the field that said
// Computed.
func check(pass *analysis.Pass, cl *ast.CompositeLit, visited map[*ast.CompositeLit]bool, where string, flags bool) {
	visited[cl] = true
	kvs := keyValues(cl)

	if flags {
		for _, name := range []string{"Optional", "Required"} {
			kv := kvs[name]
			if kv == nil || !astx.IsTrueConstant(pass, kv.Value) {
				continue
			}
			d := analysis.Diagnostic{
				Pos:     kv.Key.Pos(),
				Message: fmt.Sprintf("%s has no effect %s - use Computed", name, where),
			}
			// in Attributes() the field belongs in Arguments() instead, which is a person's call
			if where != inAttributes {
				if edit, ok := flagFix(pass, kvs, kv); ok {
					d.SuggestedFixes = []analysis.SuggestedFix{{Message: "Replace with Computed", TextEdits: []analysis.TextEdit{edit}}}
				}
			}
			pass.Report(d)
		}
	}

	for _, name := range inputOnly {
		kv := kvs[name]
		if kv == nil || !isSet(pass, kv.Value) {
			continue
		}
		pass.Report(analysis.Diagnostic{
			Pos:            kv.Key.Pos(),
			Message:        fmt.Sprintf("%s has no effect %s - remove it", name, where),
			SuggestedFixes: []analysis.SuggestedFix{{Message: "Remove " + name, TextEdits: []analysis.TextEdit{astx.DeleteLine(pass, kv)}}},
		})
	}
	if kv := kvs["ConfigMode"]; kv != nil && selectorName(kv.Value) == "SchemaConfigModeBlock" {
		pass.Report(analysis.Diagnostic{
			Pos:            kv.Key.Pos(),
			Message:        fmt.Sprintf("ConfigMode: SchemaConfigModeBlock has no effect %s - remove it", where),
			SuggestedFixes: []analysis.SuggestedFix{{Message: "Remove ConfigMode", TextEdits: []analysis.TextEdit{astx.DeleteLine(pass, kv)}}},
		})
	}

	elem, ok := stripAddr(value(kvs, "Elem")).(*ast.CompositeLit)
	if !ok {
		return
	}
	switch {
	case tf.IsSchemaHelperType(pass, elem, "Schema"):
		check(pass, elem, visited, onElem, false)
	case tf.IsSchemaHelperType(pass, elem, "Resource"):
		props, ok := value(keyValues(elem), "Schema").(*ast.CompositeLit)
		if !ok {
			return
		}
		for _, elt := range props.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if prop, ok := stripAddr(kv.Value).(*ast.CompositeLit); ok && tf.IsSchemaHelperType(pass, prop, "Schema") {
				check(pass, prop, visited, inBlock, true)
			}
		}
	}
}

// flagFix turns a nested field's Optional/Required into Computed: the key is renamed when the
// literal has no Computed yet, and the line goes when it already does.
func flagFix(pass *analysis.Pass, kvs map[string]*ast.KeyValueExpr, kv *ast.KeyValueExpr) (analysis.TextEdit, bool) {
	if kvs["Computed"] != nil {
		return astx.DeleteLine(pass, kv), true
	}
	if id, ok := kv.Value.(*ast.Ident); !ok || id.Name != "true" {
		return analysis.TextEdit{}, false
	}
	return analysis.TextEdit{Pos: kv.Key.Pos(), End: kv.Key.End(), NewText: []byte("Computed")}, true
}

// inAttributesMethod reports whether the stack sits inside a method named Attributes returning
// a map of schema pointers: azurerm's typed SDK marks every entry Computed at start-up.
func inAttributesMethod(pass *analysis.Pass, stack []ast.Node) bool {
	for _, n := range slices.Backward(stack) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name != "Attributes" || fn.Recv == nil || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
			return false
		}
		m, ok := types.Unalias(pass.TypesInfo.TypeOf(fn.Type.Results.List[0].Type)).(*types.Map)
		if !ok {
			return false
		}
		ptr, ok := types.Unalias(m.Elem()).(*types.Pointer)
		if !ok {
			return false
		}
		named, ok := types.Unalias(ptr.Elem()).(*types.Named)
		return ok && named.Obj().Name() == "Schema" && named.Obj().Pkg() != nil && named.Obj().Pkg().Name() == "schema"
	}
	return false
}

// isSet reports whether e gives the attribute a value: not nil, and not a zero constant.
func isSet(pass *analysis.Pass, e ast.Expr) bool {
	tv := pass.TypesInfo.Types[e]
	if tv.IsNil() {
		return false
	}
	if tv.Value == nil {
		return true
	}
	switch tv.Value.Kind() {
	case constant.Bool:
		return constant.BoolVal(tv.Value)
	case constant.String:
		return constant.StringVal(tv.Value) != ""
	case constant.Int, constant.Float:
		return constant.Sign(tv.Value) != 0
	case constant.Complex, constant.Unknown:
		return true
	}
	return true
}

func keyValues(cl *ast.CompositeLit) map[string]*ast.KeyValueExpr {
	kvs := make(map[string]*ast.KeyValueExpr, len(cl.Elts))
	for _, elt := range cl.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok {
				kvs[key.Name] = kv
			}
		}
	}
	return kvs
}

func value(kvs map[string]*ast.KeyValueExpr, name string) ast.Expr {
	if kv := kvs[name]; kv != nil {
		return kv.Value
	}
	return nil
}

func stripAddr(e ast.Expr) ast.Expr {
	if u, ok := e.(*ast.UnaryExpr); ok && u.Op == token.AND {
		return u.X
	}
	return e
}

func selectorName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.SelectorExpr:
		return v.Sel.Name
	case *ast.Ident:
		return v.Name
	}
	return ""
}
