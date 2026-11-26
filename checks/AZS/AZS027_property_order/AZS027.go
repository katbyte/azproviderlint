// Package AZS027 defines an analyzer that reports a resource whose well-known properties are not where a reader looks for them.
package AZS027

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks where a resource's schema puts the properties every resource has: what names it and where it lives come first, and tags come
// last, so every resource starts and ends the same way.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS027",
	Doc:      "check that name, resource_group_name and location come first in a resource's schema, and tags last",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS027_property_order/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// first are the properties that open a schema, in order, and last the ones that close it. Each default is replaced by the first one given.
var first, last, givenFirst, givenLast = []string{"name", "resource_group_name", "location"}, []string{"tags"}, false, false

// parents are what the name of a property ends in when it says what the resource sits in, which is part of its ID and may sit among the first.
var parents, givenParents = []string{"_name", "_id"}, false

func init() {
	Analyzer.Flags.Init("AZS027", flag.ContinueOnError)
	Analyzer.Flags.Func("first", "a property that goes first, in the order they are given; repeat for each, which replaces the default of name, resource_group_name and location", func(name string) error {
		if !givenFirst {
			first, givenFirst = nil, true
		}
		first = append(first, name)

		return nil
	})
	Analyzer.Flags.Func("parent-ending", "what the name of a property that may sit among the first ends in; repeat for each, which replaces the default of _name and _id", func(ending string) error {
		if !givenParents {
			parents, givenParents = nil, true
		}
		parents = append(parents, ending)

		return nil
	})
	Analyzer.Flags.Func("last", "a property that goes last; repeat for each, which replaces the default of tags", func(name string) error {
		if !givenLast {
			last, givenLast = nil, true
		}
		last = append(last, name)

		return nil
	})
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, in whatever order it was written
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	// properties reports whether a value is a map of properties written out: plugin SDK schemas, or framework attributes or blocks
	properties := func(n ast.Node) bool {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || pass.TypesInfo.TypeOf(lit) == nil {
			return false
		}

		of, isMap := pass.TypesInfo.TypeOf(lit).Underlying().(*types.Map)
		if !isMap {
			return false
		}

		elem := of.Elem()
		if ptr, isPointer := types.Unalias(elem).(*types.Pointer); isPointer {
			elem = ptr.Elem()
		}

		named, isNamed := types.Unalias(elem).(*types.Named)

		return isNamed && named.Obj().Pkg() != nil && named.Obj().Pkg().Name() == "schema" && (named.Obj().Name() == "Schema" || strings.HasSuffix(named.Obj().Name(), "Attribute") || strings.HasSuffix(named.Obj().Name(), "Block"))
	}

	insp.WithStack([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		lit, isLit := n.(*ast.CompositeLit)
		if !push || !isLit || !properties(n) || tf.InDataSourceFile(pass, n) {
			return true
		}

		// only the resource's own properties: a block inside it is its own small thing, and what is only read back is kept in Attributes
		for _, outer := range stack[:len(stack)-1] {
			if decl, isDecl := outer.(*ast.FuncDecl); properties(outer) || (isDecl && decl.Name.Name == "Attributes") {
				return true
			}
		}

		// each property by name, and whether it is written out as one that is only read back
		var names []string
		var keys []ast.Expr
		var readOnly []bool
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			name, isName := pass.TypesInfo.Types[kv.Key]
			if !isName || name.Value == nil || name.Value.Kind() != constant.String {
				continue
			}

			value := kv.Value
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}

			set, computed := false, false
			if property, isProperty := value.(*ast.CompositeLit); isProperty {
				for _, field := range property.Elts {
					if pair, isField := field.(*ast.KeyValueExpr); isField && astx.IsTrueConstant(pass, pair.Value) {
						set, computed = set || types.ExprString(pair.Key) == "Required" || types.ExprString(pair.Key) == "Optional", computed || types.ExprString(pair.Key) == "Computed"
					}
				}
			}

			names, keys, readOnly = append(names, constant.StringVal(name.Value)), append(keys, kv.Key), append(readOnly, computed && !set)
		}

		for i, name := range names {
			switch {
			case slices.Contains(first, name):
				// only the first properties ranked before it, and what the resource sits in, may be written above it
				for _, above := range names[:i] {
					if rank := slices.Index(first, above); rank > slices.Index(first, name) || (rank < 0 && !slices.ContainsFunc(parents, func(ending string) bool { return strings.HasSuffix(above, ending) })) {
						pass.Reportf(keys[i].Pos(), "property %q should come before %q - %s go first", name, above, strings.Join(first, ", "))

						break
					}
				}

			case slices.Contains(last, name):
				// what is only read back is listed after everything the user sets
				for j, below := range names[i+1:] {
					if !slices.Contains(last, below) && !readOnly[i+1+j] {
						pass.Reportf(keys[i].Pos(), "property %q should come after %q - it goes last", name, below)

						break
					}
				}
			}
		}

		return true
	})

	return nil, nil
}
