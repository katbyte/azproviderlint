// Package AZS004 defines an analyzer that reports enum validations built from a hand-written
// []string that does not cover every possible value of the SDK enum being validated.
package AZS004

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks `validation.StringInSlice([]string{...}, ...)` calls whose slice references
// an SDK enum's constants. The only accepted form is the SDK's own possible-values helper
// (`PossibleValuesFor<Enum>()` / `Possible<Enum>Values()`): a partial hand-written list
// silently rejects values the API accepts, and even a complete one goes stale the moment the
// SDK adds a new value, so both are reported — with the missing values named when the list is
// incomplete.
//
// A type only counts as an enum when its package exports such a helper, so ordinary named
// string types with a few convenience constants are not reported. Slices containing
// non-constant elements or constants of more than one enum type are skipped, since coverage
// of a deliberate union or a computed list cannot be proven statically.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS004",
	Doc:      "check that enum validation uses the SDK's possible-values helper or lists every possible value",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS004_enum_validation_possible_values/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// The two allow flags suppress one reporting class each: lists that are a deliberate subset
// of the enum (allow-missing-values) or a deliberate superset carrying legacy extras
// (allow-extra-values). A list that is exactly the enum is always reported, since switching
// to the helper is a pure win there.
var (
	allowMissingValues bool
	allowExtraValues   bool
)

func init() {
	Analyzer.Flags.BoolVar(&allowMissingValues, "allow-missing-values", false, "do not report in-place validation arrays that are missing enum values (deliberate subsets)")
	Analyzer.Flags.BoolVar(&allowExtraValues, "allow-extra-values", false, "do not report in-place validation arrays containing values that are not part of the enum (deliberate supersets)")
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		validationPkg := stringInSliceValidationPkg(pass, call)
		if validationPkg == nil {
			return
		}

		// only a hand-written []string literal can be proven complete or incomplete; calls
		// (e.g. PossibleValuesFor<Enum>()) and variables are out of scope
		lit, ok := ast.Unparen(call.Args[0]).(*ast.CompositeLit)
		if !ok {
			return
		}

		covered := map[string]bool{}
		var coveredOrder []string
		var enums []*types.Named
		for _, elem := range lit.Elts {
			tv, ok := pass.TypesInfo.Types[elem]
			if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
				// a non-constant element may contribute any value, so coverage is unprovable
				return
			}
			if v := constant.StringVal(tv.Value); !covered[v] {
				covered[v] = true
				coveredOrder = append(coveredOrder, v)
			}
			enums = appendEnumConstTypes(pass, enums, elem)
		}

		// no enum constants referenced (plain strings), or a deliberate union of two enums
		if len(enums) != 1 {
			return
		}

		enum := enums[0]
		helper, typed, values := enumValues(enum)
		if helper == "" {
			return
		}

		var missing []string
		for _, v := range values {
			if !covered[v.value] {
				missing = append(missing, fmt.Sprintf("%s (%q)", v.name, v.value))
			}
		}

		isValue := map[string]bool{}
		for _, v := range values {
			isValue[v.value] = true
		}

		var extra []string
		for _, v := range coveredOrder {
			if !isValue[v] {
				extra = append(extra, strconv.Quote(v))
			}
		}

		pkgName := enum.Obj().Pkg().Name()
		helperExpr := fmt.Sprintf("%s.%s()", pkgName, helper)
		if typed {
			// track-1 style helpers return []Enum, which StringInSlice does not accept, so
			// the advice must convert. When the matched validation package exports a generic
			// enum-slice wrapper (func StringInEnumSlice[T ~string]([]T, bool), e.g. azurerm's
			// internal/tf/validation), advise calling that directly; otherwise bridge with
			// go-azure-helpers' generic enum-slice conversion.
			if hasStringInEnumSlice(validationPkg) {
				helperExpr = fmt.Sprintf("%s.StringInEnumSlice(%s, %s)", validationPkg.Name(), helperExpr, types.ExprString(call.Args[1]))
			} else {
				helperExpr = fmt.Sprintf("pointer.FromEnumSlice(pointer.To(%s))", helperExpr)
			}
		}

		if len(missing) == 0 && len(extra) == 0 {
			pass.Reportf(lit.Pos(), "enum validation for %s.%s lists every value manually; use %s so new values are picked up automatically", pkgName, enum.Obj().Name(), helperExpr)
			return
		}

		var clauses []string
		if len(missing) > 0 && !allowMissingValues {
			clauses = append(clauses, "is missing "+strings.Join(missing, ", "))
		}
		if len(extra) > 0 && !allowExtraValues {
			clauses = append(clauses, "has extra values not in the enum: "+strings.Join(extra, ", "))
		}
		if len(clauses) == 0 {
			return
		}

		advice := "use " + helperExpr
		if len(extra) > 0 {
			// a plain swap to the helper would drop the extras, so the advice must keep them
			advice += ", appending any deliberate extras"
		}

		pass.Reportf(lit.Pos(), "enum validation for %s.%s %s; %s", pkgName, enum.Obj().Name(), strings.Join(clauses, " and "), advice)
	})

	return nil, nil
}

// stringInSliceValidationPkg returns the package declaring the StringInSlice helper the call
// resolves to, or nil when the call is not one: a function named StringInSlice with a
// ([]string, bool) parameter list, declared in a package whose import path is or ends in
// "validation". This matches the plugin SDK's helper/validation package and provider-internal
// wrappers of it (e.g. azurerm's internal/tf/validation) without depending on what the
// wrapper package is called locally.
func stringInSliceValidationPkg(pass *analysis.Pass, call *ast.CallExpr) *types.Package {
	if len(call.Args) < 2 {
		return nil
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	obj := pass.TypesInfo.ObjectOf(sel.Sel)
	if obj == nil || obj.Name() != "StringInSlice" {
		return nil
	}

	pkg := obj.Pkg()
	if pkg == nil || (pkg.Path() != "validation" && !strings.HasSuffix(pkg.Path(), "/validation")) {
		return nil
	}

	sig, ok := obj.Type().(*types.Signature)
	if !ok || sig.Params().Len() != 2 {
		return nil
	}

	slice, ok := sig.Params().At(0).Type().(*types.Slice)
	if !ok {
		return nil
	}

	elem, ok := slice.Elem().(*types.Basic)
	if !ok || elem.Kind() != types.String {
		return nil
	}

	basic, ok := sig.Params().At(1).Type().(*types.Basic)
	if !ok || basic.Kind() != types.Bool {
		return nil
	}

	return pkg
}

// hasStringInEnumSlice reports whether the validation package exports a generic enum-slice
// wrapper the typed-helper advice can name instead of the go-azure-helpers conversion:
// a function StringInEnumSlice[T ~string](valid []T, ignoreCase bool) like the one azurerm's
// internal/tf/validation gained alongside the track-1 call-site migration.
func hasStringInEnumSlice(pkg *types.Package) bool {
	fn, ok := pkg.Scope().Lookup("StringInEnumSlice").(*types.Func)
	if !ok {
		return false
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.TypeParams().Len() != 1 || sig.Params().Len() != 2 {
		return false
	}

	slice, ok := sig.Params().At(0).Type().(*types.Slice)
	if !ok {
		return false
	}

	if _, isTypeParam := types.Unalias(slice.Elem()).(*types.TypeParam); !isTypeParam {
		return false
	}

	basic, ok := sig.Params().At(1).Type().(*types.Basic)
	return ok && basic.Kind() == types.Bool
}

// appendEnumConstTypes walks elem and appends the named string type of every enum constant it
// references (deduplicated), handling `string(pkg.Const)` conversions, parenthesised forms and
// bare constant references uniformly.
func appendEnumConstTypes(pass *analysis.Pass, enums []*types.Named, elem ast.Expr) []*types.Named {
	ast.Inspect(elem, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		c, ok := pass.TypesInfo.Uses[id].(*types.Const)
		if !ok {
			return true
		}

		named, ok := types.Unalias(c.Type()).(*types.Named)
		if !ok {
			return true
		}

		basic, ok := named.Underlying().(*types.Basic)
		if !ok || basic.Info()&types.IsString == 0 {
			return true
		}

		for _, seen := range enums {
			if seen.Obj() == named.Obj() {
				return true
			}
		}

		enums = append(enums, named)
		return true
	})

	return enums
}

type enumValue struct {
	name  string
	value string
}

// enumValues returns the name of the possible-values helper for the enum, whether it returns
// a typed slice, and every constant of the enum type declared in its package, in
// declaration-scope order. The helper's presence is what marks the type as a closed enum:
// the generated go-azure-sdk packages emit `PossibleValuesFor<Name>() []string`, and the
// older track-1 style SDKs emit `Possible<Name>Values() []<Name>` — the latter returns
// typed true, since the helper cannot be passed to StringInSlice directly and the advice
// must go through go-azure-helpers' enum-slice conversion. A named string type without a
// qualifying helper returns ("", false, nil) and is not treated as an enum.
func enumValues(named *types.Named) (helper string, typed bool, values []enumValue) {
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil {
		return "", false, nil
	}

	for _, candidate := range []string{"PossibleValuesFor" + obj.Name(), "Possible" + obj.Name() + "Values"} {
		fn, ok := pkg.Scope().Lookup(candidate).(*types.Func)
		if !ok {
			continue
		}

		sig, ok := fn.Type().(*types.Signature)
		if !ok || sig.Params().Len() != 0 || sig.Results().Len() != 1 {
			continue
		}

		slice, ok := sig.Results().At(0).Type().(*types.Slice)
		if !ok {
			continue
		}

		switch {
		case types.Identical(slice.Elem(), types.Typ[types.String]):
			typed = false
		case types.Identical(types.Unalias(slice.Elem()), named):
			typed = true
		default:
			continue
		}

		helper = candidate
		break
	}

	if helper == "" {
		return "", false, nil
	}

	for _, name := range pkg.Scope().Names() {
		c, ok := pkg.Scope().Lookup(name).(*types.Const)
		if !ok || !types.Identical(types.Unalias(c.Type()), named) {
			continue
		}
		values = append(values, enumValue{name: name, value: constant.StringVal(c.Val())})
	}

	return helper, typed, values
}
