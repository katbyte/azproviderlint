// Package AZS010 defines an analyzer that reports resources, properties and blocks that have no description.
package AZS010

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that everything the docs describe has a description to describe it with: each resource and data source, each property and
// each block. The docs are made from the descriptions, so one left out is a gap on the page.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS010",
	Doc:      "check that every resource, property and block has a description",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS010_description_missing/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const dataSource = "data source"

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// a state migration's schema is what old state looked like, and no docs are made from it
	if tf.InStateMigration(pass) {
		return nil, nil
	}

	insp.WithStack([]ast.Node{(*ast.CompositeLit)(nil), (*ast.TypeSpec)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push || strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return true
		}

		// a typed resource or data source describes itself with a Description() method
		if spec, ok := n.(*ast.TypeSpec); ok {
			obj, ok := pass.TypesInfo.Defs[spec.Name].(*types.TypeName)
			if !ok || types.IsInterface(obj.Type()) {
				return true
			}

			// a field of that name, such as the one on the acceptance tests' data, does not make a resource
			named, _, _ := types.LookupFieldOrMethod(obj.Type(), true, pass.Pkg, "ResourceType")
			if _, method := named.(*types.Func); !method {
				return true
			}

			if described, _, _ := types.LookupFieldOrMethod(obj.Type(), true, pass.Pkg, "Description"); described == nil {
				what := "resource"
				if tf.InDataSourceFile(pass, spec) {
					what = dataSource
				}
				pass.Reportf(spec.Name.Pos(), "%s %s has no Description() method", what, obj.Name())
			}

			return true
		}

		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		t := pass.TypesInfo.TypeOf(cl)
		if t == nil {
			return true
		}
		if ptr, pointer := types.Unalias(t).(*types.Pointer); pointer {
			t = ptr.Elem()
		}

		// the plugin sdk and the framework keep their schema types in a package named schema, and a provider function's in one named function
		named, ok := types.Unalias(t).(*types.Named)
		if !ok || named.Obj().Pkg() == nil || !strings.Contains(named.Obj().Pkg().Path(), "/terraform-plugin-") {
			return true
		}

		name, ofFunction := named.Obj().Name(), named.Obj().Pkg().Name() == "function"
		st, ok := named.Underlying().(*types.Struct)
		if !ok || (named.Obj().Pkg().Name() != "schema" && !ofFunction) {
			return true
		}

		// only what can hold a description can be missing one
		describable, hasElem := false, false
		for field := range st.Fields() {
			describable = describable || field.Name() == "Description" || field.Name() == "MarkdownDescription"
			hasElem = hasElem || field.Name() == "Elem"
		}

		fields := map[string]ast.Expr{}
		reads := false
		for _, elt := range cl.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if key, ok := kv.Key.(*ast.Ident); ok {
					fields[key.Name] = kv.Value
					reads = reads || strings.HasPrefix(key.Name, "Read")
				}
			}
		}

		if !describable || fields["Description"] != nil || fields["MarkdownDescription"] != nil {
			return true
		}

		// where the value sits, past the & that takes its address
		at := len(stack) - 2
		for at >= 0 {
			if _, ok := stack[at].(*ast.UnaryExpr); !ok {
				break
			}
			at--
		}

		key := ""
		pos := cl.Pos()
		if at >= 0 {
			switch parent := stack[at].(type) {
			case *ast.KeyValueExpr:
				// the element of a list or a set is not a property of its own
				if ident, ok := parent.Key.(*ast.Ident); ok && ident.Name == "Elem" {
					return true
				}

				if tv, ok := pass.TypesInfo.Types[parent.Key]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
					key, pos = fmt.Sprintf(" %q", constant.StringVal(tv.Value)), parent.Key.Pos()
				}

			case *ast.CallExpr:
				// a schema handed to a function, as to the one that hashes a set, is not a property either
				return true
			}
		}

		what := "property"
		switch {
		case ofFunction && name == "Definition":
			what = "function"

		case ofFunction:
			what = "parameter"

		case name == "Resource":
			// a resource or data source is the one that can be read; the others hold a block's properties, or old state
			if !reads {
				return true
			}

			what = "resource"
			if tf.InDataSourceFile(pass, cl) {
				what = dataSource
			}

		case name == "Schema" && !hasElem:
			what = "schema"

		case strings.HasSuffix(name, "Block"):
			what = "block"

		case fields["Elem"] != nil:
			elem := pass.TypesInfo.TypeOf(fields["Elem"])
			if ptr, ok := types.Unalias(elem).(*types.Pointer); ok {
				elem = ptr.Elem()
			}
			if of, ok := types.Unalias(elem).(*types.Named); ok && of.Obj().Name() == "Resource" {
				what = "block"
			}
		}

		pass.Reportf(pos, "%s%s has no description", what, key)

		return true
	})

	return nil, nil
}
