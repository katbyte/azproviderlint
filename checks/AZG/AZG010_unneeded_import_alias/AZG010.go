// Package AZG010 reports import aliases that are not needed because the package's own name
// would not clash with anything in the file.
package AZG010

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks each aliased import for whether the package's own name could be used
// instead: nothing else in the file (another import, a package-level declaration, a local
// variable, parameter or type) and no builtin goes by that name. The alias then only
// renames the package, or repeats its name, and can be dropped. Package-level declarations
// include those of the package's own test files, which are read from disk since a pass over
// the package without its tests does not carry them.
//
// A package whose name differs from what its import path suggests (`devices` imported from
// `.../iothub`) keeps an explicit name, as goimports writes it; the fix replaces the alias
// with the package name rather than removing it, and such an explicit name is not reported.
// Two imports with the same package name are both left alone, even when both are aliased,
// since dropping either alias is only safe while the other one stays.
var Analyzer = &analysis.Analyzer{
	Name: "AZG010",
	Doc:  "check for import aliases that are not needed because the package name does not clash",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG010_unneeded_import_alias/README.md",
	Run:  run,
}

// ignore skips imports whose package name matches any of these, such as versioned SDK
// packages (`v2021_03_01`) that read better under a descriptive alias.
var ignore []*regexp.Regexp

// allow lists aliases that are never reported, whatever package they name.
var allow []string

// checkGenerated includes generated files, where a report means the generator needs fixing:
// those with the standard `// Code generated ... DO NOT EDIT.` header, and the provider's
// _gen.go files, which carry a header of their own.
var checkGenerated bool

// both flags add to their list each time they are set, since a regular expression cannot be
// split on a separator that it may itself contain
func init() {
	Analyzer.Flags.Init("AZG010", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&checkGenerated, "generated", true,
		"check generated files (false skips them)")
	Analyzer.Flags.Func("ignore", "skip imports whose package name matches this regular expression; repeat for several",
		func(s string) error {
			pattern, err := regexp.Compile(s)
			if err != nil {
				return err
			}
			ignore = append(ignore, pattern)
			return nil
		})
	Analyzer.Flags.Func("allow", "alias to leave unreported; repeat or comma-separate for several",
		func(s string) error {
			for alias := range strings.SplitSeq(s, ",") {
				allow = append(allow, strings.TrimSpace(alias))
			}
			return nil
		})
}

func run(pass *analysis.Pass) (any, error) {
	// the main package go test writes to drive a package's tests imports it as _test and
	// _xtest; it lives in the build cache, not the source tree
	if pass.Pkg.Name() == "main" && strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	// package-level names declared by the package's own test files, built on first need
	var testDeclared map[string]bool

	for _, file := range pass.Files {
		if !checkGenerated && (ast.IsGenerated(file) || strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_gen.go")) {
			continue
		}

		// names declared anywhere in the file, built on first need: renaming an import to one
		// of them would shadow it, or be shadowed, somewhere
		var declared map[string]bool

		for _, spec := range file.Imports {
			if spec.Name == nil || spec.Name.Name == "_" || spec.Name.Name == "." {
				continue
			}
			pkgName, ok := pass.TypesInfo.Defs[spec.Name].(*types.PkgName)
			if !ok {
				continue
			}
			alias := spec.Name.Name
			name := pkgName.Imported().Name()

			// the name goimports assumes from the import path; when the package is named
			// otherwise it writes the name explicitly, so the fix keeps one too
			base := path.Base(pkgName.Imported().Path())
			if version, ok := strings.CutPrefix(base, "v"); ok {
				if _, err := strconv.Atoi(version); err == nil {
					base = path.Base(path.Dir(pkgName.Imported().Path()))
				}
			}
			base = strings.TrimPrefix(base, "go-")
			if i := strings.IndexFunc(base, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' }); i >= 0 {
				base = base[:i]
			}
			explicit := base != name

			if alias == name && explicit {
				continue
			}
			if slices.Contains(allow, alias) {
				continue
			}
			if slices.ContainsFunc(ignore, func(pattern *regexp.Regexp) bool { return pattern.MatchString(name) }) {
				continue
			}
			if types.Universe.Lookup(name) != nil || pass.Pkg.Scope().Lookup(name) != nil {
				continue
			}

			clash := false
			for _, other := range file.Imports {
				if other == spec {
					continue
				}
				var obj types.Object
				if other.Name != nil {
					obj = pass.TypesInfo.Defs[other.Name]
				} else {
					obj = pass.TypesInfo.Implicits[other]
				}
				if other, ok := obj.(*types.PkgName); ok && (other.Name() == name || other.Imported().Name() == name) {
					clash = true
				}
			}
			if clash {
				continue
			}

			if declared == nil {
				declared = map[string]bool{}
				ast.Inspect(file, func(n ast.Node) bool {
					id, ok := n.(*ast.Ident)
					if !ok {
						return true
					}
					obj, defines := pass.TypesInfo.Defs[id]
					switch obj := obj.(type) {
					case *types.Var:
						declared[id.Name] = declared[id.Name] || !obj.IsField()
					case *types.Func:
						declared[id.Name] = declared[id.Name] || obj.Signature().Recv() == nil
					case *types.Const, *types.TypeName:
						declared[id.Name] = true
					case nil:
						// the variable of a type switch (`switch x := v.(type)`) has no object
						// of its own, and neither does the package clause, which is not a clash
						declared[id.Name] = declared[id.Name] || (defines && id != file.Name)
					}
					return true
				})
			}
			if declared[name] {
				continue
			}

			if testDeclared == nil {
				testDeclared = map[string]bool{}
				dir := filepath.Dir(pass.Fset.Position(file.Pos()).Filename)
				entries, _ := os.ReadDir(dir)
				for _, entry := range entries {
					if !strings.HasSuffix(entry.Name(), "_test.go") {
						continue
					}
					testFile, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
					if err != nil || testFile.Name.Name != pass.Pkg.Name() {
						continue
					}
					for _, decl := range testFile.Decls {
						switch decl := decl.(type) {
						case *ast.FuncDecl:
							if decl.Recv == nil {
								testDeclared[decl.Name.Name] = true
							}
						case *ast.GenDecl:
							for _, spec := range decl.Specs {
								switch spec := spec.(type) {
								case *ast.TypeSpec:
									testDeclared[spec.Name.Name] = true
								case *ast.ValueSpec:
									for _, valueName := range spec.Names {
										testDeclared[valueName.Name] = true
									}
								}
							}
						}
					}
				}
			}
			if testDeclared[name] {
				continue
			}

			message := fmt.Sprintf("import alias %q is not needed, the package name %q does not clash", alias, name)
			fixMessage := "Use the package name instead of the alias"
			if alias == name {
				message = fmt.Sprintf("import alias %q repeats the package name", alias)
				fixMessage = "Drop the alias"
			}

			edits := []analysis.TextEdit{{Pos: spec.Name.Pos(), End: spec.Path.Pos()}}
			if explicit {
				edits[0] = analysis.TextEdit{Pos: spec.Name.Pos(), End: spec.Name.End(), NewText: []byte(name)}
			}
			if alias != name {
				ast.Inspect(file, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == pkgName {
						edits = append(edits, analysis.TextEdit{Pos: id.Pos(), End: id.End(), NewText: []byte(name)})
					}
					return true
				})
			}

			pass.Report(analysis.Diagnostic{
				Pos:            spec.Pos(),
				End:            spec.End(),
				Message:        message,
				SuggestedFixes: []analysis.SuggestedFix{{Message: fixMessage, TextEdits: edits}},
			})
		}
	}
	return nil, nil
}
