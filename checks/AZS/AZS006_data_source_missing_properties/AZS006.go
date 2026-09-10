// Package AZS006 defines an analyzer that reports data sources whose same-named resource
// exposes schema properties the data source does not, across untyped plugin SDK, typed SDK
// and framework registration flavours.
package AZS006

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"slices"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/katbyte/azproviderlint/checks/azignore"
)

// Analyzer pairs every registered data source with the same-named registered resource and
// compares their schema property names. Properties declared anywhere in the resource's schema
// (top-level or nested) that appear nowhere in the data source's schema are reported on the
// data source's registration entry.
//
// Schema property trees are built recursively: string-keyed map literals whose value type is
// a schema type (`*pluginsdk.Schema`, framework `Attribute`/`Block`) contribute their
// constant keys, nested Elem schemas become children, and calls into same-package functions
// (schema helper functions, shared schema builders) are followed. A resource property counts
// as covered if its name appears anywhere in the data source's schema — this under-reports on
// restructured schemas but never false-positives from restructuring. A missing block is
// reported once, by dotted path, with its entire subtree suppressed. Pairs where either side
// yields no properties, or where the data source contains a non-constant key, are skipped.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS006",
	Doc:      "check for data sources missing properties that exist on the corresponding resource",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS006_data_source_missing_properties/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// ignoreSensitive drops resource properties marked `Sensitive: true` from the comparison.
// Some providers deliberately keep secrets out of data sources, where values land in state
// readable by anyone with state access; the flag makes that policy checkable.
var ignoreSensitive bool

func init() {
	Analyzer.Flags.BoolVar(&ignoreSensitive, "ignore-sensitive", false,
		"do not report resource properties marked Sensitive: true")
}

var (
	resourceMethods   = map[string]bool{"SupportedResources": true, "Resources": true, "FrameworkResources": true}
	dataSourceMethods = map[string]bool{"SupportedDataSources": true, "DataSources": true, "FrameworkDataSources": true}
)

// entry is one registered resource or data source: its terraform type name, the registration
// position, and where its schema lives — an expression (untyped map value, e.g.
// `dataSourceFoo()`) or a named type (typed/framework slice element).
type entry struct {
	name     string
	pos      token.Pos
	expr     ast.Expr
	typeName string
}

type collector struct {
	pass    *analysis.Pass
	decls   map[types.Object]*ast.FuncDecl
	methods map[string]map[string]*ast.FuncDecl // receiver type name -> method name -> decl
	ignored map[string]map[int]bool             // filename -> lines carrying //azignore:AZS006
}

// ignoredAt reports whether pos falls on (or directly below) an //azignore:AZS006 directive.
func (c *collector) ignoredAt(pos token.Pos) bool {
	p := c.pass.Fset.Position(pos)
	return c.ignored[p.Filename][p.Line]
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	c := &collector{pass: pass, decls: map[types.Object]*ast.FuncDecl{}, methods: map[string]map[string]*ast.FuncDecl{}, ignored: azignore.Lines(pass, "AZS006")}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if obj := pass.TypesInfo.Defs[fn.Name]; obj != nil {
				c.decls[obj] = fn
			}
			if fn.Recv != nil && len(fn.Recv.List) == 1 {
				recv := fn.Recv.List[0].Type
				if star, ok := recv.(*ast.StarExpr); ok {
					recv = star.X
				}
				if id, ok := recv.(*ast.Ident); ok {
					if c.methods[id.Name] == nil {
						c.methods[id.Name] = map[string]*ast.FuncDecl{}
					}
					c.methods[id.Name][fn.Name.Name] = fn
				}
			}
		}
	}

	var resources, dataSources []entry

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
		if (!isResource && !isDataSource) || !tf.RegistrationReturnShape(pass, fn) {
			return
		}

		entries := c.collectRegistrations(fn.Body)
		if isResource {
			resources = append(resources, entries...)
		} else {
			dataSources = append(dataSources, entries...)
		}
	})

	resourceByName := map[string]entry{}
	for _, r := range resources {
		if _, ok := resourceByName[r.name]; !ok {
			resourceByName[r.name] = r
		}
	}

	reported := map[string]bool{}
	for _, ds := range dataSources {
		res, ok := resourceByName[ds.name]
		if !ok || reported[ds.name] {
			continue
		}
		reported[ds.name] = true

		// per-property //azignore:AZS006 directives are honoured on the resource side only:
		// an ignored resource property is simply never required of the data source
		resTree, _ := c.schemaTree(res, true)
		dsTree, dsClean := c.schemaTree(ds, false)
		if len(resTree.children) == 0 || len(dsTree.children) == 0 || !dsClean {
			continue
		}

		dsNames := map[string]bool{}
		dsTree.allNames(dsNames)

		var missing []string
		resTree.walkMissing(dsNames, "", &missing)
		if len(missing) == 0 {
			continue
		}
		slices.Sort(missing)

		pass.Reportf(ds.pos, "data source %q is missing resource properties: %s", ds.name, strings.Join(missing, ", "))
	}

	return nil, nil
}

// isWriteOnly reports whether a property name follows the provider's write-only argument
// convention (`foo_wo` plus its `foo_wo_version` companion). Write-only arguments are never
// readable, so a data source cannot expose them by definition.
func isWriteOnly(name string) bool {
	return strings.HasSuffix(name, "_wo") || strings.HasSuffix(name, "_wo_version")
}

// collectRegistrations walks a registration method body collecting each registered entry with
// the location of its schema: map literal entries and `m[key] = value` assignments keep the
// value expression, slice literal elements and append arguments keep the element's type name.
func (c *collector) collectRegistrations(body *ast.BlockStmt) []entry {
	var entries []entry

	addTyped := func(elem ast.Expr) {
		t := types.Unalias(c.pass.TypesInfo.TypeOf(elem))
		if ptr, ok := t.(*types.Pointer); ok {
			t = types.Unalias(ptr.Elem())
		}
		named, ok := t.(*types.Named)
		if !ok {
			return
		}
		typeName := named.Obj().Name()
		name, ok := c.resourceTypeOf(typeName)
		if !ok {
			return
		}
		entries = append(entries, entry{name: name, pos: elem.Pos(), typeName: typeName})
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			t := types.Unalias(c.pass.TypesInfo.TypeOf(node))
			if _, isMap := t.(*types.Map); isMap {
				for _, elt := range node.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					if name, ok := c.constantString(kv.Key); ok {
						entries = append(entries, entry{name: name, pos: kv.Key.Pos(), expr: kv.Value})
					}
				}
				return false
			}
			if _, isSlice := t.(*types.Slice); isSlice {
				for _, elt := range node.Elts {
					addTyped(elt)
				}
				return false
			}
		case *ast.AssignStmt:
			for i, lhs := range node.Lhs {
				idx, ok := lhs.(*ast.IndexExpr)
				if !ok || i >= len(node.Rhs) {
					continue
				}
				if _, isMap := types.Unalias(c.pass.TypesInfo.TypeOf(idx.X)).(*types.Map); !isMap {
					continue
				}
				if name, ok := c.constantString(idx.Index); ok {
					entries = append(entries, entry{name: name, pos: idx.Index.Pos(), expr: node.Rhs[i]})
				}
			}
		case *ast.CallExpr:
			if fun, ok := node.Fun.(*ast.Ident); ok && fun.Name == "append" && len(node.Args) > 1 {
				for _, arg := range node.Args[1:] {
					if c.delegatesToRegistrationMethod(arg) {
						continue
					}
					addTyped(arg)
				}
				return false
			}
		}
		return true
	})

	return entries
}

// treeNode is one level of a schema: property name -> the nested schema beneath it.
type treeNode struct {
	children  map[string]*treeNode
	writeOnly bool
	sensitive bool
}

func newTreeNode() *treeNode {
	return &treeNode{children: map[string]*treeNode{}}
}

func (t *treeNode) child(name string) *treeNode {
	if t.children[name] == nil {
		t.children[name] = newTreeNode()
	}
	return t.children[name]
}

// allNames flattens every property name at every level into out.
func (t *treeNode) allNames(out map[string]bool) {
	for name, child := range t.children {
		out[name] = true
		child.allNames(out)
	}
}

// walkMissing reports resource properties whose name appears nowhere in the data source. A
// missing property suppresses its entire subtree — reporting a missing block's every nested
// property adds nothing once the block itself is called out. Nested gaps under covered
// parents are reported with their dotted path.
func (t *treeNode) walkMissing(dsNames map[string]bool, prefix string, missing *[]string) {
	for name, child := range t.children {
		if child.writeOnly || isWriteOnly(name) {
			continue
		}
		if ignoreSensitive && child.sensitive {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if !dsNames[name] {
			*missing = append(*missing, strconv.Quote(path))
			continue
		}
		child.walkMissing(dsNames, path, missing)
	}
}

// schemaTree resolves the property tree for a registration entry, and whether every
// encountered schema key was constant. Typed/framework entries use the type's
// Arguments()+Attributes() methods when present, falling back to its Schema() method;
// untyped entries follow the registration value expression into its function declaration.
func (c *collector) schemaTree(e entry, honorIgnores bool) (*treeNode, bool) {
	root := newTreeNode()
	b := &treeBuilder{c: c, clean: true, honorIgnores: honorIgnores}

	if e.typeName != "" {
		methods := c.methods[e.typeName]
		if fromMethods := methods["Arguments"] != nil || methods["Attributes"] != nil; fromMethods {
			for _, name := range []string{"Arguments", "Attributes"} {
				if m := methods[name]; m != nil && m.Body != nil {
					b.build(m.Body, root)
				}
			}
		} else if m := methods["Schema"]; m != nil && m.Body != nil {
			b.build(m.Body, root)
		}
		return root, b.clean
	}

	if fn := c.declOfCall(e.expr); fn != nil && fn.Body != nil {
		b.build(fn.Body, root)
	}
	return root, b.clean
}

// treeBuilder builds property trees, tracking the call stack to guard against recursive
// schema helpers and whether every key seen was constant.
type treeBuilder struct {
	c            *collector
	stack        []*ast.FuncDecl
	clean        bool
	honorIgnores bool
}

func (b *treeBuilder) onStack(fn *ast.FuncDecl) bool {
	return slices.Contains(b.stack, fn)
}

// build walks a node collecting schema-map keys into the given level. Schema map literals
// found at this level contribute their keys here, with each key's value walked one level
// down; calls into same-package functions stay at the current level, so whole-map helper
// functions and merge patterns land where they are used.
func (b *treeBuilder) build(n ast.Node, node *treeNode) {
	ast.Inspect(n, func(x ast.Node) bool {
		switch v := x.(type) {
		case *ast.CompositeLit:
			m, ok := types.Unalias(b.c.pass.TypesInfo.TypeOf(v)).(*types.Map)
			if !ok || !isSchemaMap(m) {
				return true
			}
			b.collectMap(v, node)
			return false
		case *ast.AssignStmt:
			// m["prop"] = ... on a schema map adds a property at this level; any other
			// assignment (s := map{...}, r := &Resource{...}) is walked normally
			handled := false
			for i, lhs := range v.Lhs {
				idx, ok := lhs.(*ast.IndexExpr)
				if !ok || i >= len(v.Rhs) {
					continue
				}
				m, ok := types.Unalias(b.c.pass.TypesInfo.TypeOf(idx.X)).(*types.Map)
				if !ok || !isSchemaMap(m) {
					continue
				}
				handled = true
				if name, ok := b.c.constantString(idx.Index); ok {
					if b.honorIgnores && b.c.ignoredAt(idx.Index.Pos()) {
						continue
					}
					child := node.child(name)
					if b.hasBoolField(v.Rhs[i], "WriteOnly") {
						child.writeOnly = true
					}
					if b.hasBoolField(v.Rhs[i], "Sensitive") {
						child.sensitive = true
					}
					b.build(v.Rhs[i], child)
				} else {
					b.clean = false
				}
			}
			return !handled
		case *ast.CallExpr:
			if fn := b.c.declOfCall(v); fn != nil && fn.Body != nil && !b.onStack(fn) {
				b.stack = append(b.stack, fn)
				b.build(fn.Body, node)
				b.stack = b.stack[:len(b.stack)-1]
			}
			return true // still walk arguments: merge(map1, map2) carries maps at this level
		}
		return true
	})
}

// hasBoolField reports whether a property's own schema literal sets the named bool field to
// true (e.g. `WriteOnly: true`, `Sensitive: true`). Descent stops at nested schema maps,
// whose properties carry their own markers.
func (b *treeBuilder) hasBoolField(expr ast.Expr, field string) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if lit, ok := n.(*ast.CompositeLit); ok {
			if m, ok := types.Unalias(b.c.pass.TypesInfo.TypeOf(lit)).(*types.Map); ok && isSchemaMap(m) {
				return false
			}
		}
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		if id, isIdent := kv.Key.(*ast.Ident); isIdent && id.Name == field {
			if tv, hasType := b.c.pass.TypesInfo.Types[kv.Value]; hasType && tv.Value != nil && tv.Value.Kind() == constant.Bool && constant.BoolVal(tv.Value) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// collectMap adds a schema map literal's keys at the given level and walks each value one
// level down, so nested Elem schemas become children rather than siblings.
func (b *treeBuilder) collectMap(lit *ast.CompositeLit, node *treeNode) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name, ok := b.c.constantString(kv.Key)
		if !ok {
			b.clean = false
			continue
		}
		if b.honorIgnores && b.c.ignoredAt(kv.Key.Pos()) {
			continue
		}
		child := node.child(name)
		if b.hasBoolField(kv.Value, "WriteOnly") {
			child.writeOnly = true
		}
		if b.hasBoolField(kv.Value, "Sensitive") {
			child.sensitive = true
		}
		b.build(kv.Value, child)
	}
}

// isSchemaMap reports whether m is a schema property map: string keys with a value type named
// Schema (plugin SDK, possibly behind a pointer) or Attribute/Block (framework).
func isSchemaMap(m *types.Map) bool {
	basic, ok := types.Unalias(m.Key()).(*types.Basic)
	if !ok || basic.Kind() != types.String {
		return false
	}

	elem := types.Unalias(m.Elem())
	if ptr, isPtr := elem.(*types.Pointer); isPtr {
		elem = types.Unalias(ptr.Elem())
	}
	named, ok := elem.(*types.Named)
	if !ok {
		return false
	}
	switch named.Obj().Name() {
	case "Schema", "Attribute", "Block":
		return true
	}
	return false
}

// declOfCall resolves a call expression to the declaration of the same-package function or
// method it invokes.
func (c *collector) declOfCall(expr ast.Expr) *ast.FuncDecl {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok {
		return nil
	}

	var id *ast.Ident
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		id = fun
	case *ast.SelectorExpr:
		id = fun.Sel
	default:
		return nil
	}

	obj, ok := c.pass.TypesInfo.ObjectOf(id).(*types.Func)
	if !ok || obj.Pkg() != c.pass.Pkg {
		return nil
	}

	if fn := c.decls[obj]; fn != nil {
		return fn
	}
	// methods are declared under their receiver in Defs too, but fall back to the index
	if recv := obj.Signature().Recv(); recv != nil {
		t := types.Unalias(recv.Type())
		if ptr, ok := t.(*types.Pointer); ok {
			t = types.Unalias(ptr.Elem())
		}
		if named, ok := t.(*types.Named); ok {
			return c.methods[named.Obj().Name()][obj.Name()]
		}
	}
	return nil
}

// delegatesToRegistrationMethod reports whether expr is a call to another registration method
// (`r.autoRegistration.DataSources()` etc.), whose entries are collected independently.
func (c *collector) delegatesToRegistrationMethod(expr ast.Expr) bool {
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
	_, isFunc := c.pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	return isFunc
}

// resourceTypeOf resolves a type name to the constant string its ResourceType() method
// returns.
func (c *collector) resourceTypeOf(typeName string) (string, bool) {
	decl := c.methods[typeName]["ResourceType"]
	if decl == nil || decl.Body == nil {
		return "", false
	}

	var name string
	found := false
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || found || len(ret.Results) != 1 {
			return true
		}
		name, found = c.constantString(ret.Results[0])
		return !found
	})

	return name, found
}

// constantString resolves expr to a constant string via the type checker, falling back to a
// package-level `var` with a constant string initializer.
func (c *collector) constantString(expr ast.Expr) (string, bool) {
	if tv, ok := c.pass.TypesInfo.Types[expr]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
		return constant.StringVal(tv.Value), true
	}

	id, ok := ast.Unparen(expr).(*ast.Ident)
	if !ok {
		return "", false
	}
	v, ok := c.pass.TypesInfo.ObjectOf(id).(*types.Var)
	if !ok || v.Pkg() != c.pass.Pkg || v.Parent() != c.pass.Pkg.Scope() {
		return "", false
	}

	for _, file := range c.pass.Files {
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
					if tv, ok := c.pass.TypesInfo.Types[vs.Values[i]]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
						return constant.StringVal(tv.Value), true
					}
					return "", false
				}
			}
		}
	}

	return "", false
}
