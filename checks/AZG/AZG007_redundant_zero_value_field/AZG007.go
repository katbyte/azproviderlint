// Package AZG007 defines an analyzer that reports redundant zero-value assignments to struct
// literal fields, where the field should be omitted instead.
package AZG007

import (
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/checks/azignore"
	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
)

const ruleName = "AZG007"

// Analyzer checks for fields explicitly initialised to their zero value in a struct literal —
// a pointer set to `nil`, a string set to `""`, a numeric set to `0`, or a bool set to `false` —
// which is redundant because an omitted field already takes its zero value. Only pointer and
// basic (string/numeric/bool) fields are flagged; slices, maps, and interfaces are left alone
// because an explicit nil there can be a deliberate, readable signal. Zero values written as a
// named constant (`Type: TypeNone`) are also left alone, since the name conveys intent. Test
// files are skipped by default (the tests flag opts in), since a zero entry in a test table is
// often semantically meaningful.
var Analyzer = &analysis.Analyzer{
	Name: ruleName,
	Doc:  "check for redundant zero-value assignments to struct literal fields that should be omitted",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG007_redundant_zero_value_field/README.md",
	Run:  run,
}

// checkTests opts into reporting inside test files, which are skipped by default: a zero entry
// in a test table is often a deliberate, meaningful row rather than redundant noise.
var checkTests bool

func init() {
	Analyzer.Flags.Init("AZG007", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&checkTests, "tests", false,
		"also report in test files, where a zero entry in a test table is often meaningful (off by default)")
}

func run(pass *analysis.Pass) (any, error) {
	ignored := azignore.ExactLines(pass, ruleName)

	for _, file := range pass.Files {
		// A zero entry in a test table is often semantically meaningful, so skip test files
		// unless the tests flag opts in.
		if !checkTests && strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}

		// This file's comment groups, position-sorted, used to avoid deleting a comment that
		// leads the field following the one being removed.
		commentGroups := file.Comments

		ast.Inspect(file, func(n ast.Node) bool {
			compositeLit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			pos := pass.Fset.Position(compositeLit.Pos())
			if ignored[pos.Filename][pos.Line] {
				return false
			}

			for i, elt := range compositeLit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				keyIdent, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}

				field, ok := pass.TypesInfo.Uses[keyIdent].(*types.Var)
				if !ok || !field.IsField() {
					continue
				}

				message := redundantZeroMessage(pass, field.Type(), kv.Value, keyIdent.Name)
				if message == "" {
					continue
				}

				diagnostic := analysis.Diagnostic{Pos: kv.Pos(), Message: message}

				// A standalone comment directly above the field often explains why it is set;
				// deleting the field would re-attach it to the next one, so report without a
				// fix and let a person decide (or convert the comment to an azignore).
				prevEnd := compositeLit.Lbrace
				if i > 0 {
					prevEnd = compositeLit.Elts[i-1].End()
				}
				prevLine := pass.Fset.Position(prevEnd).Line
				kvLine := pass.Fset.Position(kv.Pos()).Line
				leadComment := false
				for _, cg := range commentGroups {
					if cg.Pos() > prevEnd && cg.End() < kv.Pos() &&
						pass.Fset.Position(cg.Pos()).Line > prevLine &&
						pass.Fset.Position(cg.End()).Line == kvLine-1 {
						leadComment = true
						break
					}
				}

				if !leadComment {
					// Delete the element up to the start of the next one (or the closing brace
					// for the last element), which takes the trailing comma and any trailing
					// comment with it; gofmt collapses the leftover whitespace afterwards.
					end := compositeLit.Rbrace
					if i+1 < len(compositeLit.Elts) {
						end = compositeLit.Elts[i+1].Pos()
					}

					// A comment on a later line than this field leads the following field, so
					// stop the deletion before it; a same-line trailing comment stays inside
					// the span.
					eltLine := pass.Fset.Position(elt.End()).Line
					for _, cg := range commentGroups {
						if cg.Pos() > elt.End() && cg.Pos() < end && pass.Fset.Position(cg.Pos()).Line > eltLine {
							end = cg.Pos()
							break
						}
					}

					diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
						Message:   "Remove the redundant field",
						TextEdits: []analysis.TextEdit{{Pos: kv.Pos(), End: end}},
					}}
				}

				pass.Report(diagnostic)
			}

			return true
		})
	}

	return nil, nil
}

// redundantZeroMessage returns a diagnostic message when value is a redundant zero-value
// assignment for a field of fieldType, or an empty string otherwise. Pointer fields are
// redundant when set to `nil`; string/numeric/bool fields when set to a literal `""`, `0`, or
// `false`. Slices, maps, interfaces, and named-constant zero values are left alone.
func redundantZeroMessage(pass *analysis.Pass, fieldType types.Type, value ast.Expr, name string) string {
	switch underlying := fieldType.Underlying().(type) {
	case *types.Pointer:
		if astx.IsNilValue(pass, value) {
			return fmt.Sprintf("redundant nil assignment to pointer field %q - omit the field", name)
		}
	case *types.Basic:
		if isRedundantZero(pass, underlying, value) {
			return fmt.Sprintf("redundant zero-value assignment to field %q - omit the field", name)
		}
	}
	return ""
}

// isRedundantZero reports whether value is a compile-time zero for the basic field type — `0`,
// `-0.0`, `'\x00'`, `int64(0)`, `""`, `false` — written without naming a constant. Any
// reference to a named (non-predeclared) constant keeps the field: a zero spelled `TypeNone`
// documents intent, and omitting the field would lose that.
func isRedundantZero(pass *analysis.Pass, basic *types.Basic, value ast.Expr) bool {
	cv := pass.TypesInfo.Types[value].Value
	if cv == nil || cv.Kind() == constant.Unknown {
		return false
	}

	named := false
	ast.Inspect(value, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if c, ok := pass.TypesInfo.Uses[id].(*types.Const); ok && c.Pkg() != nil {
				named = true
			}
		}
		return !named
	})
	if named {
		return false
	}

	switch info := basic.Info(); {
	case info&types.IsBoolean != 0:
		return cv.Kind() == constant.Bool && !constant.BoolVal(cv)
	case info&types.IsString != 0:
		return cv.Kind() == constant.String && constant.StringVal(cv) == ""
	case info&types.IsComplex != 0:
		return constant.Sign(constant.Real(cv)) == 0 && constant.Sign(constant.Imag(cv)) == 0
	case info&types.IsNumeric != 0:
		return constant.Sign(cv) == 0
	}
	return false
}
