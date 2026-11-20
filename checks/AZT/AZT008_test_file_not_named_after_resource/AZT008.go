// Package AZT008 defines an analyzer that reports acceptance tests in a file that is not named
// after the file declaring the resource or data source they test.
package AZT008

import (
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/astx"
)

// Analyzer checks that a test sits in a file named after the file declaring what it tests:
// tests for the resource declared in `foo_resource.go` belong in `foo_resource_test.go` or,
// unless the suffix option is off, `foo_resource_<part>_test.go`. What a test tests is what it
// builds test data for, the second argument of `acceptance.BuildTestData`.
var Analyzer = &analysis.Analyzer{
	Name: "AZT008",
	Doc:  "check that acceptance tests are in a file named after the file declaring the resource they test",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT008_test_file_not_named_after_resource/README.md",
	Run:  run,
}

// suffix accepts `foo_resource_<part>_test.go` as well as `foo_resource_test.go`, so one
// resource's tests can be split across files.
var suffix = true

func init() {
	Analyzer.Flags.Init("AZT008", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&suffix, "suffix", true, "also accept a test file with a part between the declaring file's name and _test.go (false requires exactly <file>_test.go)")
}

// What a registration method registers, as reports name it.
const (
	resource   = "resource"
	dataSource = "data source"
)

var registers = map[string]string{
	"SupportedResources":   resource,
	"Resources":            resource,
	"FrameworkResources":   resource,
	"SupportedDataSources": dataSource,
	"DataSources":          dataSource,
	"FrameworkDataSources": dataSource,
}

func run(pass *analysis.Pass) (any, error) {
	// the functions in test files that build test data, with what each passes as the type
	type test struct {
		file       string
		decl       *ast.FuncDecl
		built      []ast.Expr
		dataSource bool // also hands t to a data source harness
	}

	var tests []test
	dir := ""
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if !strings.HasSuffix(filename, "_test.go") {
			continue
		}

		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}

			tt := test{file: filepath.Base(filename), decl: fd}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				fn := astx.CalledFunc(pass, call)
				if fn == nil || fn.Pkg() == nil || !strings.HasSuffix(fn.Pkg().Path(), "/internal/acceptance") {
					return true
				}

				switch {
				case fn.Name() == "BuildTestData" && len(call.Args) > 1:
					tt.built = append(tt.built, call.Args[1])
				case strings.HasPrefix(fn.Name(), "DataSource"):
					tt.dataSource = true
				}

				return true
			})

			if len(tt.built) > 0 {
				tests = append(tests, tt)
				dir = filepath.Dir(filename)
			}
		}
	}

	if len(tests) == 0 {
		return nil, nil
	}

	// Test files are a separate package from the one they test, so that package's own files are
	// read from disk. Syntax is enough to tell what it registers: a registered value is a call
	// to one of its functions, or a literal of one of its types that ResourceType names.
	names, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	funcFile, typeFile := map[string]string{}, map[string]string{} // the file declaring each function and type
	strs := map[string]string{}                                    // package-level string constants and variables
	resourceType := map[string]ast.Expr{}                          // what each type's ResourceType method returns
	var registrations []*ast.FuncDecl
	for _, entry := range names {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						typeFile[s.Name.Name] = entry.Name()
					case *ast.ValueSpec:
						for i, name := range s.Names {
							if i >= len(s.Values) {
								break
							}
							if lit, ok := s.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
								strs[name.Name], _ = strconv.Unquote(lit.Value)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil {
					funcFile[d.Name.Name] = entry.Name()
					continue
				}
				if d.Body == nil {
					continue
				}

				if registers[d.Name.Name] != "" {
					registrations = append(registrations, d)
				}

				recv := d.Recv.List[0].Type
				if star, ok := recv.(*ast.StarExpr); ok {
					recv = star.X
				}
				if id, ok := recv.(*ast.Ident); ok && d.Name.Name == "ResourceType" && len(d.Body.List) == 1 {
					if ret, ok := d.Body.List[0].(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
						resourceType[id.Name] = ret.Results[0]
					}
				}
			}
		}
	}

	// a terraform type name: a string literal, or a package-level constant or variable
	text := func(e ast.Expr) string {
		switch e := e.(type) {
		case *ast.BasicLit:
			if s, err := strconv.Unquote(e.Value); err == nil && e.Kind == token.STRING {
				return s
			}
		case *ast.Ident:
			return strs[e.Name]
		}
		return ""
	}

	type declared struct{ kind, name, file string }

	byName, byType := map[string]declared{}, map[string]declared{}
	for _, registration := range registrations {
		kind := registers[registration.Name.Name]
		ast.Inspect(registration.Body, func(n ast.Node) bool {
			var key, value ast.Expr
			switch n := n.(type) {
			case *ast.KeyValueExpr: // a map entry: the name, then a call to the declaring function
				key, value = n.Key, n.Value
			case *ast.AssignStmt: // the same assigned by index, as behind a feature flag
				if index, ok := n.Lhs[0].(*ast.IndexExpr); ok && len(n.Rhs) == 1 {
					key, value = index.Index, n.Rhs[0]
				}
			case *ast.CompositeLit: // a typed or framework element, named by its ResourceType method
				if id, ok := n.Type.(*ast.Ident); ok && typeFile[id.Name] != "" && text(resourceType[id.Name]) != "" {
					d := declared{kind, text(resourceType[id.Name]), typeFile[id.Name]}
					byName[kind+" "+d.name], byType[id.Name] = d, d
				}
			}

			if call, ok := value.(*ast.CallExpr); ok {
				if fn, ok := call.Fun.(*ast.Ident); ok && funcFile[fn.Name] != "" && text(key) != "" {
					byName[kind+" "+text(key)] = declared{kind, text(key), funcFile[fn.Name]}
				}
			}

			return true
		})
	}

	// A test is in the right file when the file is named after the declaring file of anything
	// it builds test data for. Otherwise it is reported against the first, once per test file.
	type misplaced struct {
		declared
		first *ast.FuncDecl
		count int
	}

	found := map[string]*misplaced{}
	var order []string
	own := strings.TrimSuffix(pass.Pkg.Path(), "_test")
	for _, tt := range tests {
		var want []declared
		for _, arg := range tt.built {
			// "azurerm_foo", or "data.azurerm_foo" for a data source. A test run as a data source
			// test under the bare name may be either, so both files are right for it.
			if v := pass.TypesInfo.Types[arg].Value; v != nil && v.Kind() == constant.String {
				name, isData := strings.CutPrefix(constant.StringVal(v), "data.")
				if !isData {
					want = append(want, byName[resource+" "+name])
				}
				if isData || tt.dataSource {
					want = append(want, byName[dataSource+" "+name])
				}

				continue
			}

			// FooResource{}.ResourceType(), on its own or inside a fmt.Sprintf("data.%s", ...)
			ast.Inspect(arg, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "ResourceType" {
					return true
				}

				t := pass.TypesInfo.TypeOf(sel.X)
				if ptr, ok := t.(*types.Pointer); ok {
					t = ptr.Elem()
				}
				if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == own {
					want = append(want, byType[named.Obj().Name()])
				}

				return true
			})
		}

		// nothing registered under that name in this package: not this check's to judge
		want = slices.DeleteFunc(want, func(d declared) bool { return d.file == "" })
		right := slices.ContainsFunc(want, func(d declared) bool {
			stem := strings.TrimSuffix(d.file, ".go")
			return tt.file == stem+"_test.go" || (suffix && strings.HasPrefix(tt.file, stem+"_"))
		})
		if len(want) == 0 || right {
			continue
		}

		key := tt.file + " " + want[0].kind + " " + want[0].name
		if found[key] == nil {
			found[key] = &misplaced{declared: want[0], first: tt.decl}
			order = append(order, key)
		}

		found[key].count++
	}

	for _, key := range order {
		m := found[key]
		stem := strings.TrimSuffix(m.file, ".go")
		who, belongs := m.first.Name.Name+" builds", "it belongs"
		if m.count > 1 {
			who, belongs = fmt.Sprintf("%s and %d more in this file build", m.first.Name.Name, m.count-1), "they belong"
		}

		where := stem + "_test.go"
		if suffix {
			where += " or " + stem + "_<part>_test.go"
		}

		pass.Reportf(m.first.Name.Pos(), "%s test data for %s %q, declared in %s, so %s in %s", who, m.kind, m.name, m.file, belongs, where)
	}

	return nil, nil
}
