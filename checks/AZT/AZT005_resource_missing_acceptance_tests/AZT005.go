// Package AZT005 defines an analyzer that reports registered resources and data sources
// missing the acceptance tests every one of them should have.
package AZT005

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
)

// Analyzer checks that every registered data source has a `basic` acceptance test and every
// registered resource has `basic` and `requiresImport` ones, plus `complete` and `update`
// when the resource can be updated in place. A resource can be updated when it declares an
// Update handler: the plugin SDK requires one exactly when a field is not ForceNew.
//
// The tests live in the package's `_test.go` files, which are a separate package, so they
// are read from disk: every `TestAcc`/`testAcc` function in a test file named after the
// file declaring the resource, or whose name spells the resource's type name, counts.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT005",
	Doc:      "check that resources and data sources have basic, requiresImport, complete and update acceptance tests",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT005_resource_missing_acceptance_tests/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// The registration keys that give an untyped resource an Update handler.
var updateKeys = map[string]bool{"Update": true, "UpdateContext": true, "UpdateWithoutTimeout": true}

// testFunc is one acceptance test function found in a test file: TestAccFoo_list_basic has
// prefix Foo and segments [list basic].
type testFunc struct {
	file     string // base name
	prefix   string
	segments []string
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	funcDecls := map[*types.Func]*ast.FuncDecl{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Body != nil {
				if fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func); ok {
					funcDecls[fn] = fd
				}
			}
		}
	}

	type registered struct {
		tf.Entry
		dataSource bool
	}
	var entries []registered
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Body == nil {
			return
		}
		isResource := tf.ResourceMethods[fn.Name.Name]
		isDataSource := tf.DataSourceMethods[fn.Name.Name]
		if (!isResource && !isDataSource) || !tf.RegistrationReturnShape(pass, fn) {
			return
		}
		found, _ := tf.RegistrationEntries(pass, fn.Body)
		for _, e := range found {
			entries = append(entries, registered{Entry: e, dataSource: isDataSource})
		}
	})

	testsByDir := map[string][]testFunc{}
	seen := map[string]bool{}
	for _, e := range entries {
		kind := "resource"
		if e.dataSource {
			kind = "data source"
		}
		if seen[kind+e.Name] || e.Expr == nil {
			continue
		}
		seen[kind+e.Name] = true

		file, updatable, ok := declaration(pass, funcDecls, e.Expr)
		if !ok {
			continue
		}
		dir := filepath.Dir(file)
		tests, cached := testsByDir[dir]
		if !cached {
			tests = acceptanceTests(dir)
			testsByDir[dir] = tests
		}

		steps := []string{"basic"}
		if !e.dataSource {
			steps = append(steps, "requiresImport")
			if updatable {
				steps = append(steps, "complete", "update")
			}
		}

		// the tests for this entry: in a test file named after the declaring file, or named
		// after the type name itself
		stem := strings.TrimSuffix(filepath.Base(file), ".go") + "_"
		typeName := strings.TrimPrefix(e.Name, "azurerm_")
		prefix := ""
		covered := map[string]bool{}
		for _, t := range tests {
			if !strings.HasPrefix(t.file, stem) && !spells(t.prefix, typeName, e.dataSource) {
				continue
			}
			if prefix == "" {
				prefix = t.prefix
			}
			for _, seg := range t.segments {
				for _, step := range steps {
					if strings.HasPrefix(strings.ToLower(seg), strings.ToLower(step)) {
						covered[step] = true
					}
				}
			}
		}
		if prefix == "" {
			prefix = camel(typeName)
		}

		var missing []string
		for _, step := range steps {
			if !covered[step] {
				missing = append(missing, "TestAcc"+prefix+"_"+step)
			}
		}
		if len(missing) == 0 {
			continue
		}
		pass.Reportf(e.Pos, "%s %q is missing acceptance tests: %s", kind, e.Name, strings.Join(missing, ", "))
	}

	return nil, nil
}

// declaration resolves a registered value to the file declaring the resource and whether the
// resource has an Update handler. An untyped `resourceFoo()` call resolves to the function,
// whose schema.Resource literal is read for an Update key; a typed or framework `FooResource{}`
// resolves to the type, which has an Update method or not. Anything declared outside the
// package is not resolved.
func declaration(pass *analysis.Pass, funcDecls map[*types.Func]*ast.FuncDecl, expr ast.Expr) (file string, updatable, ok bool) {
	expr = ast.Unparen(expr)
	if call, isCall := expr.(*ast.CallExpr); isCall {
		fd := funcDecls[astx.CalledFunc(pass, call)]
		if fd == nil {
			return "", false, false
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			cl, isLit := n.(*ast.CompositeLit)
			if !isLit || !tf.IsSchemaHelperType(pass, cl, "Resource") {
				return true
			}
			for _, elt := range cl.Elts {
				kv, isKV := elt.(*ast.KeyValueExpr)
				if !isKV {
					continue
				}
				if key, isIdent := kv.Key.(*ast.Ident); isIdent && updateKeys[key.Name] && !astx.IsNilValue(pass, kv.Value) {
					updatable = true
				}
			}
			return true
		})
		return pass.Fset.Position(fd.Pos()).Filename, updatable, true
	}

	t := types.Unalias(pass.TypesInfo.TypeOf(expr))
	if ptr, isPtr := t.(*types.Pointer); isPtr {
		t = types.Unalias(ptr.Elem())
	}
	named, isNamed := t.(*types.Named)
	if !isNamed || named.Obj().Pkg() != pass.Pkg {
		return "", false, false
	}
	update, _, _ := types.LookupFieldOrMethod(named, true, pass.Pkg, "Update")
	_, updatable = update.(*types.Func)
	return pass.Fset.Position(named.Obj().Pos()).Filename, updatable, true
}

// acceptanceTests parses every _test.go file in dir and returns its TestAcc/testAcc
// functions. Test files belong to a separate package, so they are read from disk rather than
// from the pass. A file that does not parse contributes nothing.
func acceptanceTests(dir string) []testFunc {
	names, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var tests []testFunc
	fset := token.NewFileSet()
	for _, entry := range names {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil {
				continue
			}
			rest, ok := strings.CutPrefix(fd.Name.Name, "TestAcc")
			if !ok {
				rest, ok = strings.CutPrefix(fd.Name.Name, "testAcc")
			}
			if !ok {
				continue
			}
			parts := strings.Split(rest, "_")
			tests = append(tests, testFunc{file: entry.Name(), prefix: parts[0], segments: parts[1:]})
		}
	}
	return tests
}

// spells reports whether a test name prefix spells the terraform type name (without its
// provider prefix), ignoring case and underscores: TestAccMsSqlServer for mssql_server. A
// leading AzureRM, as older tests carry, and a Resource suffix are allowed; a data source
// may carry DataSource at either end.
func spells(prefix, typeName string, dataSource bool) bool {
	p := strings.TrimPrefix(strings.ToLower(prefix), "azurerm")
	want := strings.ReplaceAll(typeName, "_", "")
	if dataSource {
		return p == want || p == "datasource"+want || p == want+"datasource"
	}
	return p == want || p == want+"resource"
}

// camel turns foo_bar into FooBar, the test name prefix a new resource would use.
func camel(typeName string) string {
	var b strings.Builder
	upper := true
	for _, r := range typeName {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			r = unicode.ToUpper(r)
			upper = false
		}
		b.WriteRune(r)
	}
	return b.String()
}
