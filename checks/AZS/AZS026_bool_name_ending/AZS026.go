// Package AZS026 defines an analyzer that reports a bool property whose name does not end in _enabled or a word for its state.
package AZS026

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks how a bool property is named, so a user can tell from the name alone that it is on or off: compression_enabled, or a word
// for its state where _enabled reads oddly, such as mtls_required.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS026",
	Doc:      "check that a bool property ends in _enabled, or a word for its state such as _required",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS026_bool_name_ending/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// endings are what a bool's name may end in: _enabled, and the words for a state that azurerm's naming guide says read better without it. The
// default is replaced by the first one given.
var endings, givenEndings = []string{"_enabled", "_required", "_accepted", "_allowed", "_supported", "_verified", "_configured", "_only"}, false

func init() {
	Analyzer.Flags.Init("AZS026", flag.ContinueOnError)
	Analyzer.Flags.Func("ending", "what a bool's name may end in, such as _enabled; repeat for each, which replaces the default", func(ending string) error {
		if !givenEndings {
			endings, givenEndings = nil, true
		}
		endings = append(endings, ending)

		return nil
	})
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, and its names cannot change
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	// a property is named by its key in the map it sits in, which is met before the property is
	keys := map[*ast.CompositeLit]ast.Expr{}
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || pass.TypesInfo.TypeOf(lit) == nil {
			return
		}

		fields := map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := kv.Value
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}

			if inner, isInner := value.(*ast.CompositeLit); isInner {
				keys[inner] = kv.Key
			}
			if key, isField := kv.Key.(*ast.Ident); isField {
				fields[key.Name] = kv.Value
			}
		}

		t := pass.TypesInfo.TypeOf(lit)
		if ptr, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
			t = ptr.Elem()
		}

		named, isNamed := types.Unalias(t).(*types.Named)
		key, isName := pass.TypesInfo.Types[keys[lit]]
		if !isNamed || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" || !isName || key.Value == nil || key.Value.Kind() != constant.String {
			return
		}

		// a name on its way out is not worth changing
		if fields["Deprecated"] != nil || fields["DeprecationMessage"] != nil {
			return
		}

		if named.Obj().Name() != "BoolAttribute" && (named.Obj().Name() != "Schema" || !strings.HasSuffix(types.ExprString(fields["Type"]), "TypeBool")) {
			return
		}

		name, ends := constant.StringVal(key.Value), false
		for _, ending := range endings {
			ends = ends || strings.HasSuffix(name, ending) || name == strings.TrimPrefix(ending, "_")
		}

		if !ends {
			allowed := strings.Join(endings, ", ")
			if last := strings.LastIndex(allowed, ", "); last >= 0 {
				allowed = allowed[:last] + " or " + allowed[last+2:]
			}
			pass.Reportf(keys[lit].Pos(), "bool property %q should end in %s", name, allowed)
		}
	})

	return nil, nil
}
