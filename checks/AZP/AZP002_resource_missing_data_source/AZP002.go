// Package AZP002 defines an analyzer that reports registered resources that have no
// corresponding data source of the same name, across every registration flavour: untyped
// plugin SDK maps, typed sdk.Resource slices and framework wrapped resource slices.
package AZP002

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer correlates a service package's registered resources and data sources by their
// terraform type name and reports every resource without a matching data source. All
// registration flavours contribute to both sides:
//
//   - untyped plugin SDK: `SupportedResources()` / `SupportedDataSources()` map keys
//   - typed SDK: `Resources()` / `DataSources()` elements' `ResourceType()` methods
//   - framework: `FrameworkResources()` / `FrameworkDataSources()` elements' `ResourceType()` methods
//
// Conditionally registered entries (`m["azurerm_x"] = ...`, `append(out, Foo{})` behind
// feature flags) are collected too. If any data source entry's name cannot be resolved the
// package is skipped entirely, so an unresolvable data source can never produce a false
// "missing data source" report.
var Analyzer = &analysis.Analyzer{
	Name:     "AZP002",
	Doc:      "check for registered resources that have no corresponding data source",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP002_resource_missing_data_source/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// registration method names for each side; the untyped map methods and the typed/framework
// slice methods are distinguished by their return shape when collecting entries.
var (
	resourceMethods   = map[string]bool{"SupportedResources": true, "Resources": true, "FrameworkResources": true}
	dataSourceMethods = map[string]bool{"SupportedDataSources": true, "DataSources": true, "FrameworkDataSources": true}
)

// actionStyleSuffixes marks invoke-style resources — ones that perform an operation or mint a
// credential rather than manage a durable object (the kind that will eventually become
// framework Actions). These have no meaningful data source form, so they are never reported.
// There is no structural marker for them in any registration flavour, so this is a name
// convention list.
var actionStyleSuffixes = []string{
	"_run_command",
	"_sas_token",
}

type entry struct {
	name string
	pos  token.Pos
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	var resources []entry
	dataSources := map[string]bool{}
	resolvable := true

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Body == nil {
			return
		}

		isResource := resourceMethods[fn.Name.Name]
		isDataSource := dataSourceMethods[fn.Name.Name]
		if !isResource && !isDataSource {
			return
		}

		if !tf.RegistrationReturnShape(pass, fn) {
			return
		}

		entries, allResolved := collectEntries(pass, fn.Body)
		if isDataSource {
			if !allResolved {
				// an unresolvable data source name could hide a match — skip the package
				resolvable = false
				return
			}
			for _, e := range entries {
				dataSources[e.name] = true
			}
			return
		}

		// unresolvable resource entries are simply skipped: they can only under-report
		resources = append(resources, entries...)
	})

	if !resolvable {
		return nil, nil
	}

	reported := map[string]bool{}
	for _, r := range resources {
		if dataSources[r.name] || reported[r.name] || isActionStyle(r.name) {
			continue
		}
		reported[r.name] = true
		pass.Reportf(r.pos, "resource %q has no corresponding data source", r.name)
	}

	return nil, nil
}

// isActionStyle reports whether a resource name marks an invoke-style resource that has no
// meaningful data source form.
func isActionStyle(name string) bool {
	for _, suffix := range actionStyleSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// collectEntries walks a registration method body and collects the terraform type name of
// every registered entry: map literal keys, `m[key] = ...` assignments, slice literal
// elements and `append(...)` arguments. It reports whether every encountered entry resolved
// to a name.
func collectEntries(pass *analysis.Pass, body *ast.BlockStmt) ([]entry, bool) {
	var entries []entry
	allResolved := true

	add := func(name string, pos token.Pos, ok bool) {
		if !ok {
			allResolved = false
			return
		}
		entries = append(entries, entry{name: name, pos: pos})
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			t := types.Unalias(pass.TypesInfo.TypeOf(node))
			if _, isMap := t.(*types.Map); isMap {
				for _, elt := range node.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					name, ok := constantString(pass, kv.Key)
					add(name, kv.Key.Pos(), ok)
				}
				return false
			}
			if _, isSlice := t.(*types.Slice); isSlice {
				for _, elt := range node.Elts {
					name, ok := resourceTypeOf(pass, elt)
					add(name, elt.Pos(), ok)
				}
				return false
			}
		case *ast.AssignStmt:
			// m["azurerm_x"] = resourceX() — conditional registration behind feature flags
			for _, lhs := range node.Lhs {
				idx, ok := lhs.(*ast.IndexExpr)
				if !ok {
					continue
				}
				if _, isMap := types.Unalias(pass.TypesInfo.TypeOf(idx.X)).(*types.Map); !isMap {
					continue
				}
				name, ok := constantString(pass, idx.Index)
				add(name, idx.Index.Pos(), ok)
			}
		case *ast.CallExpr:
			// out = append(out, FooResource{}) — conditional registration behind feature flags
			if fun, ok := node.Fun.(*ast.Ident); ok && fun.Name == "append" && len(node.Args) > 1 {
				for _, arg := range node.Args[1:] {
					// append(out, r.autoRegistration.DataSources()...) — delegation to another
					// registration method in the same package (generated auto-registration);
					// its entries are collected when that method's declaration is visited
					if delegatesToRegistrationMethod(pass, arg) {
						continue
					}
					name, ok := resourceTypeOf(pass, arg)
					add(name, arg.Pos(), ok)
				}
				return false
			}
		}
		return true
	})

	return entries, allResolved
}

// delegatesToRegistrationMethod reports whether expr is a call to another registration method
// (`r.autoRegistration.DataSources()` etc.), whose entries are collected independently from
// that method's own declaration in the package.
func delegatesToRegistrationMethod(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if !resourceMethods[name] && !dataSourceMethods[name] {
		return false
	}
	_, isFunc := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	return isFunc
}

// constantString resolves expr to a constant string via the type checker, so literals and
// named constants both work. As a fallback, a reference to a package-level `var` with a
// constant string initializer (`var FooResourceName = "azurerm_foo"`) resolves to that
// initializer, since some resources declare their type name that way for use with locks.
func constantString(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	if tv, ok := pass.TypesInfo.Types[expr]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
		return constant.StringVal(tv.Value), true
	}

	id, ok := ast.Unparen(expr).(*ast.Ident)
	if !ok {
		return "", false
	}
	v, ok := pass.TypesInfo.ObjectOf(id).(*types.Var)
	if !ok || v.Pkg() != pass.Pkg || v.Parent() != pass.Pkg.Scope() {
		return "", false
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) != len(vs.Values) {
					continue
				}
				for i, name := range vs.Names {
					if name.Name != id.Name {
						continue
					}
					if tv, ok := pass.TypesInfo.Types[vs.Values[i]]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
						return constant.StringVal(tv.Value), true
					}
					return "", false
				}
			}
		}
	}

	return "", false
}

// resourceTypeOf resolves a typed/framework registration element (`FooResource{}` or
// `&FooResource{}`) to the constant string its ResourceType() method returns.
func resourceTypeOf(pass *analysis.Pass, elem ast.Expr) (string, bool) {
	t := types.Unalias(pass.TypesInfo.TypeOf(elem))
	if ptr, ok := t.(*types.Pointer); ok {
		t = types.Unalias(ptr.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}

	decl := resourceTypeMethod(pass, named.Obj().Name())
	if decl == nil {
		return "", false
	}

	var name string
	found := false
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || found || len(ret.Results) != 1 {
			return true
		}
		name, found = constantString(pass, ret.Results[0])
		return !found
	})

	return name, found
}

// resourceTypeMethod finds the `func (r TypeName) ResourceType() string` declaration for the
// named receiver type in the package being analysed.
func resourceTypeMethod(pass *analysis.Pass, typeName string) *ast.FuncDecl {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "ResourceType" || fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Body == nil {
				continue
			}
			recv := fn.Recv.List[0].Type
			if star, ok := recv.(*ast.StarExpr); ok {
				recv = star.X
			}
			if id, ok := recv.(*ast.Ident); ok && id.Name == typeName {
				return fn
			}
		}
	}
	return nil
}
