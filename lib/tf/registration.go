// Package tf defines helpers shared by checks that reason about terraform provider code:
// plugin SDK schema literals, service registration methods and data source files.
package tf

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Registration method names for each side; the untyped map methods and the typed/framework
// slice methods are distinguished by their return shape, see RegistrationReturnShape.
var (
	ResourceMethods   = map[string]bool{"SupportedResources": true, "Resources": true, "FrameworkResources": true}
	DataSourceMethods = map[string]bool{"SupportedDataSources": true, "DataSources": true, "FrameworkDataSources": true}
)

// RegistrationReturnShape reports whether fn returns either a map[string]*Resource (untyped
// registration) or a slice of a named Resource/DataSource/FrameworkWrapped* type
// (typed/framework registration), which guards against unrelated methods sharing the names.
func RegistrationReturnShape(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}

	ret := pass.TypesInfo.TypeOf(fn.Type.Results.List[0].Type)
	switch t := types.Unalias(ret).(type) {
	case *types.Map:
		basic, ok := types.Unalias(t.Key()).(*types.Basic)
		if !ok || basic.Kind() != types.String {
			return false
		}

		ptr, ok := types.Unalias(t.Elem()).(*types.Pointer)
		if !ok {
			return false
		}

		named, ok := types.Unalias(ptr.Elem()).(*types.Named)
		return ok && named.Obj().Name() == "Resource"
	case *types.Slice:
		named, ok := types.Unalias(t.Elem()).(*types.Named)
		if !ok {
			return false
		}
		switch named.Obj().Name() {
		case "Resource", "DataSource", "FrameworkWrappedResource", "FrameworkWrappedDataSource":
			return true
		}
	}

	return false
}

// Entry is one registered resource or data source.
type Entry struct {
	Name string    // the terraform type name, "azurerm_foo"
	Pos  token.Pos // the map key or slice element that registers it
	Expr ast.Expr  // the registered value: the call in `"azurerm_foo": resourceFoo()`, the element `FooResource{}`; nil when unknown
}

// RegistrationEntries walks a registration method body and collects the terraform type
// name of every registered entry: map literal keys, `m[key] = ...` assignments, slice
// literal elements and `append(...)` arguments. Typed and framework elements are named by
// their `ResourceType()` method. It reports whether every encountered entry resolved to a
// name.
func RegistrationEntries(pass *analysis.Pass, body *ast.BlockStmt) ([]Entry, bool) {
	var entries []Entry
	allResolved := true

	add := func(name string, pos token.Pos, expr ast.Expr, ok bool) {
		if !ok {
			allResolved = false
			return
		}
		entries = append(entries, Entry{Name: name, Pos: pos, Expr: expr})
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
					add(name, kv.Key.Pos(), kv.Value, ok)
				}
				return false
			}
			if _, isSlice := t.(*types.Slice); isSlice {
				for _, elt := range node.Elts {
					name, ok := resourceTypeOf(pass, elt)
					add(name, elt.Pos(), elt, ok)
				}
				return false
			}
		case *ast.AssignStmt:
			// m["azurerm_x"] = resourceX() — conditional registration behind feature flags
			for i, lhs := range node.Lhs {
				idx, ok := lhs.(*ast.IndexExpr)
				if !ok {
					continue
				}
				if _, isMap := types.Unalias(pass.TypesInfo.TypeOf(idx.X)).(*types.Map); !isMap {
					continue
				}

				var value ast.Expr
				if len(node.Rhs) == len(node.Lhs) {
					value = node.Rhs[i]
				}

				name, ok := constantString(pass, idx.Index)
				add(name, idx.Index.Pos(), value, ok)
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
					add(name, arg.Pos(), arg, ok)
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
	if !ResourceMethods[name] && !DataSourceMethods[name] {
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
