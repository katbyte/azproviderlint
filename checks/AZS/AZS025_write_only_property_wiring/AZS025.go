// Package AZS025 defines an analyzer that reports a write-only property that is not tied to its version and to the property it stands in for.
package AZS025

import (
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

// Analyzer checks the three properties a write-only value is made of: the value itself, the version that says it has changed, and the property it
// stands in for. Terraform keeps no write-only value, so only the schema stops them being set in a way that never sends a change.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS025",
	Doc:      "check that a write-only property is tied to its version and to the property it stands in for",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS025_write_only_property_wiring/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || pass.TypesInfo.TypeOf(lit) == nil {
			return
		}

		// the properties written out in one map of them, by name
		if _, isMap := pass.TypesInfo.TypeOf(lit).Underlying().(*types.Map); !isMap {
			return
		}

		var order []string
		keys, properties := map[string]ast.Expr{}, map[string]map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := kv.Value
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}

			name, isName := pass.TypesInfo.Types[kv.Key]
			if !isName || name.Value == nil || name.Value.Kind() != constant.String {
				continue
			}

			// a property whose schema comes from a function is there, but what it says cannot be read
			keys[constant.StringVal(name.Value)] = kv.Key

			property, isProperty := value.(*ast.CompositeLit)
			if !isProperty || !tf.IsSchemaHelperType(pass, property, "Schema") {
				continue
			}

			fields := map[string]ast.Expr{}
			for _, field := range property.Elts {
				if pair, isField := field.(*ast.KeyValueExpr); isField {
					if key, isIdent := pair.Key.(*ast.Ident); isIdent {
						fields[key.Name] = pair.Value
					}
				}
			}

			order = append(order, constant.StringVal(name.Value))
			properties[constant.StringVal(name.Value)] = fields
		}

		// names reports whether one of the lists of a property names the other; a list that is not written out could name anything
		names := func(property, other string, lists ...string) bool {
			for _, list := range lists {
				given := properties[property][list]
				if given == nil {
					continue
				}

				written, isWritten := given.(*ast.CompositeLit)
				if !isWritten {
					return true
				}

				for _, elt := range written.Elts {
					if tv, isText := pass.TypesInfo.Types[elt]; isText && tv.Value != nil && tv.Value.Kind() == constant.String && (constant.StringVal(tv.Value) == other || strings.HasSuffix(constant.StringVal(tv.Value), "."+other)) {
						return true
					}
				}
			}

			return false
		}

		for _, name := range order {
			if !astx.IsTrueConstant(pass, properties[name]["WriteOnly"]) {
				continue
			}

			version := name + "_version"
			switch {
			case keys[version] == nil:
				pass.Reportf(keys[name].Pos(), "write-only property %q has no %q beside it - nothing says when its value has changed", name, version)

			case !names(name, version, "RequiredWith"):
				pass.Reportf(keys[name].Pos(), "%q can be set without %q - add RequiredWith: []string{%q}", name, version, version)
			}

			if properties[version] != nil {
				if !names(version, name, "RequiredWith") {
					pass.Reportf(keys[version].Pos(), "%q can be set without %q - add RequiredWith: []string{%q}", version, name, name)
				}

				// an unset version is stored as 0, so 0 cannot be told from no version at all
				if check := types.ExprString(properties[version]["ValidateFunc"]); !strings.Contains(check, "IntAtLeast(1)") && !strings.Contains(check, "IntBetween(1,") {
					pass.Reportf(keys[version].Pos(), "%q can be set to 0, which is what an unset version is stored as - add ValidateFunc: validation.IntAtLeast(1)", version)
				}
			}

			// the property it stands in for is the same name without _wo; one that is only read back cannot be set beside it
			stored, standsIn := strings.CutSuffix(name, "_wo")
			if !standsIn || properties[stored] == nil || (!astx.IsTrueConstant(pass, properties[stored]["Optional"]) && !astx.IsTrueConstant(pass, properties[stored]["Required"])) {
				continue
			}

			if !names(stored, name, "ConflictsWith", "ExactlyOneOf") {
				pass.Reportf(keys[stored].Pos(), "%q and %q can both be set - add ConflictsWith: []string{%q}", stored, name, name)
			}
			if !names(name, stored, "ConflictsWith", "ExactlyOneOf") {
				pass.Reportf(keys[name].Pos(), "%q and %q can both be set - add ConflictsWith: []string{%q}", name, stored, stored)
			}
		}
	})

	return nil, nil
}
