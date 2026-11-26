// Package AZS024 defines an analyzer that reports a string property the user sets that has no validation.
package AZS024

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/types"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that a string the user sets is validated, so a value the API will refuse is caught when the plan is made and not part way
// through an apply.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS024",
	Doc:      "check that string properties the user sets have validation",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS024_string_missing_validation/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// strict is whether a string that is only checked for not being empty is reported too.
var strict bool

func init() {
	Analyzer.Flags.Init("AZS024", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&strict, "strict", false, "also report a string whose only validation is that it is not empty: validation.StringIsNotEmpty, or the framework's LengthAtLeast(1)")
}

// element is what a list or set says about the schema of the values in it: the name of the property, and whether the user sets it.
type element struct {
	of  ast.Expr
	set bool
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, which nobody sets
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	// a property is named by its key in the map it sits in, and a list is met before the schema of its values is
	keys, elements := map[*ast.CompositeLit]ast.Expr{}, map[*ast.CompositeLit]element{}
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit {
			return
		}

		fields, inside := map[string]ast.Expr{}, map[string]*ast.CompositeLit{}
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := kv.Value
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}

			inner, isInner := value.(*ast.CompositeLit)
			if isInner {
				keys[inner] = kv.Key
			}
			if key, isField := kv.Key.(*ast.Ident); isField {
				fields[key.Name], inside[key.Name] = kv.Value, inner
			}
		}

		t := pass.TypesInfo.TypeOf(lit)
		if t == nil {
			return
		}
		if ptr, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
			t = ptr.Elem()
		}

		// the plugin sdk and the framework both keep their schema types in a package named schema; a property on its way out is left alone
		named, isNamed := types.Unalias(t).(*types.Named)
		if !isNamed || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" || fields["Deprecated"] != nil || fields["DeprecationMessage"] != nil {
			return
		}

		set, kind := astx.IsTrueConstant(pass, fields["Required"]) || astx.IsTrueConstant(pass, fields["Optional"]), types.ExprString(fields["Type"])
		var weak bool
		name, add := keys[lit], "a ValidateFunc that checks"
		_, isElement := elements[lit]
		switch named.Obj().Name() {
		case "StringAttribute":
			list, isList := fields["Validators"].(*ast.CompositeLit)
			weak, add = strict && isList && len(list.Elts) == 1 && strings.HasSuffix(types.ExprString(list.Elts[0]), "LengthAtLeast(1)"), "Validators that check"
			if !set || (fields["Validators"] != nil && (!isList || len(list.Elts) > 0) && !weak) {
				return
			}

		case "Schema":
			// the values of a list or set are checked one by one; those of a map cannot be
			if inside["Elem"] != nil {
				elements[inside["Elem"]] = element{of: keys[lit], set: set && !strings.HasSuffix(kind, "TypeMap")}
			}
			if isElement {
				set, name = elements[lit].set, elements[lit].of
			}

			// when strict, being told a string is not empty does not count
			check := types.ExprString(fields["ValidateFunc"])
			weak = strict && fields["ValidateDiagFunc"] == nil && (check == "StringIsNotEmpty" || strings.HasSuffix(check, ".StringIsNotEmpty"))

			validated := (fields["ValidateFunc"] != nil && !astx.IsNilValue(pass, fields["ValidateFunc"])) || (fields["ValidateDiagFunc"] != nil && !astx.IsNilValue(pass, fields["ValidateDiagFunc"]))
			if !set || (validated && !weak) || !strings.HasSuffix(kind, "TypeString") {
				return
			}

		default:
			return
		}

		what := "string property"
		if isElement {
			what = "the strings in this property"
		}
		if tv, isName := pass.TypesInfo.Types[name]; isName && tv.Value != nil && tv.Value.Kind() == constant.String {
			what = "string property " + strconv.Quote(constant.StringVal(tv.Value))
			if isElement {
				what = "the strings in " + strconv.Quote(constant.StringVal(tv.Value))
			}
		}

		wrong := "has no validation"
		switch {
		case weak && isElement:
			wrong = "are only checked for not being empty"

		case weak:
			wrong = "is only checked for not being empty"

		case isElement:
			wrong = "have no validation"
		}

		at := ast.Node(lit)
		if keys[lit] != nil {
			at = keys[lit]
		}

		pass.Reportf(at.Pos(), "%s %s - add %s what the API accepts", what, wrong, add)
	})

	return nil, nil
}
