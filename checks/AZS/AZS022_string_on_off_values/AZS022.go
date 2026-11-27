// Package AZS022 defines an analyzer that reports a string property that only takes two values meaning on and off.
package AZS022

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for a string property whose validation allows only a pair such as Enabled and Disabled, which is a bool written the long way:
// the user has to look up two words where true and false would do.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS022",
	Doc:      "check for string properties that only take two values meaning on and off, which should be a bool",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS022_string_on_off_values/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// onOff holds the pairs of values that mean on and off, in lower case: each value with the one it goes with.
var onOff = map[string]string{"enabled": "disabled", "disabled": "enabled", "enable": "disable", "disable": "enable", "on": "off", "off": "on", "true": "false", "false": "true"}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, which cannot be given another type
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	// a property is named by its key in the map it sits in, which is met before the property is
	keys := map[*ast.CompositeLit]ast.Expr{}
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit {
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

		var validators []ast.Expr
		if list, isList := fields["Validators"].(*ast.CompositeLit); isList && strings.HasSuffix(named.Obj().Name(), "Attribute") {
			validators = list.Elts
		}
		if named.Obj().Name() == "Schema" && fields["ValidateFunc"] != nil {
			validators = []ast.Expr{fields["ValidateFunc"]}
		}

		for _, validator := range validators {
			call, isCall := ast.Unparen(validator).(*ast.CallExpr)
			if !isCall || len(call.Args) == 0 {
				continue
			}

			fn := astx.CalledFunc(pass, call)
			if fn == nil || fn.Pkg() == nil {
				continue
			}

			// the values are the arguments of the framework's validator, and the first argument of the plugin sdk's
			var given []ast.Expr
			var values []string
			switch {
			case fn.Pkg().Name() == "stringvalidator" && (fn.Name() == "OneOf" || fn.Name() == "OneOfCaseInsensitive"):
				given = call.Args

			case (fn.Pkg().Path() == "validation" || strings.HasSuffix(fn.Pkg().Path(), "/validation")) && (fn.Name() == "StringInSlice" || fn.Name() == "StringInEnumSlice"):
				switch list := ast.Unparen(call.Args[0]).(type) {
				case *ast.CompositeLit:
					given = list.Elts

				case *ast.CallExpr:
					// an sdk's helper returns every value of the enum it is named for: PossibleValuesForState, or PossibleStateValues
					helper := astx.CalledFunc(pass, list)
					if helper == nil || helper.Pkg() == nil {
						continue
					}

					of, found := strings.CutPrefix(helper.Name(), "PossibleValuesFor")
					if !found {
						of, found = strings.CutPrefix(strings.TrimSuffix(helper.Name(), "Values"), "Possible")
					}

					enum, isEnum := helper.Pkg().Scope().Lookup(of).(*types.TypeName)
					if !found || !isEnum {
						continue
					}

					for _, name := range helper.Pkg().Scope().Names() {
						if c, isConst := helper.Pkg().Scope().Lookup(name).(*types.Const); isConst && types.Identical(c.Type(), enum.Type()) && c.Val().Kind() == constant.String {
							values = append(values, constant.StringVal(c.Val()))
						}
					}
				}
			}

			// a value that is not written out could be anything
			for _, value := range given {
				tv, isValue := pass.TypesInfo.Types[value]
				if !isValue || tv.Value == nil || tv.Value.Kind() != constant.String {
					values = nil
					break
				}
				values = append(values, constant.StringVal(tv.Value))
			}

			if len(values) != 2 || onOff[strings.ToLower(values[0])] == "" || onOff[strings.ToLower(values[0])] != strings.ToLower(values[1]) {
				continue
			}

			property := "string property"
			if name, isNamed := pass.TypesInfo.Types[keys[lit]]; isNamed && name.Value != nil && name.Value.Kind() == constant.String {
				property = fmt.Sprintf("string property %q", constant.StringVal(name.Value))
			}

			pass.Reportf(validator.Pos(), "%s only takes %q and %q - make it a bool", property, values[0], values[1])
		}
	})

	return nil, nil
}
