// Package AZT010 defines an analyzer that reports testdata files that no Go file names.
package AZT010

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/testdatarefs"
)

// Analyzer checks that every file in a package's testdata folder is named by a Go file in that package's folder or one beneath it. A file that
// nothing names is left over, usually from a test that was moved or deleted.
var Analyzer = &analysis.Analyzer{
	Name: "AZT010",
	Doc:  "check that the files in testdata are named by a test",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT010_testdata_file_unused/README.md",
	Run:  run,
}

// ignore skips testdata files whose path under testdata matches any of these, such as the notes on how a certificate was made.
var ignore []*regexp.Regexp

func init() {
	Analyzer.Flags.Init("AZT010", flag.ContinueOnError)
	Analyzer.Flags.Func("ignore", "skip testdata files whose path under testdata matches this regular expression; repeat for several", func(s string) error {
		pattern, err := regexp.Compile(s)
		if err != nil {
			return err
		}

		ignore = append(ignore, pattern)

		return nil
	})
}

func run(pass *analysis.Pass) (any, error) {
	anchor := anchorFile(pass)
	if anchor == nil {
		return nil, nil
	}

	dir := filepath.Dir(pass.Fset.Position(anchor.Pos()).Filename)
	testdata := filepath.Join(dir, "testdata")

	files, ok := dataFiles(testdata)
	if !ok || len(files) == 0 {
		return nil, nil
	}

	refs := mentions(dir)

	var unused []string
	for _, file := range files {
		name := filepath.ToSlash(strings.TrimPrefix(file, testdata+string(filepath.Separator)))
		if slices.ContainsFunc(ignore, func(pattern *regexp.Regexp) bool { return pattern.MatchString(name) }) {
			continue
		}

		used := slices.ContainsFunc(refs, func(ref testdatarefs.Ref) bool {
			// the end of a longer path cannot be placed, so it counts for any file of that name
			if !ref.Rooted {
				return ref.Name != "" && (name == ref.Name || strings.HasPrefix(name, ref.Name+"/"))
			}

			return ref.Covers(file)
		})
		if !used {
			unused = append(unused, name)
		}
	}

	switch len(unused) {
	case 0:
	case 1:
		pass.Reportf(anchor.Name.Pos(), "testdata/%s is not named by any Go file in this folder or beneath it", unused[0])
	default:
		pass.Reportf(anchor.Name.Pos(), "%d files in testdata are not named by any Go file in this folder or beneath it: %s", len(unused), strings.Join(unused, ", "))
	}

	return nil, nil
}

// anchorFile returns the file a folder's report goes on, or nil when this is not the pass to make it. A folder is checked as up to three packages:
// its own, its own with its tests, and its tests on their own. Picking the first file by name, tests last, gives all three the same answer.
func anchorFile(pass *analysis.Pass) *ast.File {
	if len(pass.Files) == 0 {
		return nil
	}

	dir := filepath.Dir(pass.Fset.Position(pass.Files[0].Pos()).Filename)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	first := ""
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}

		// ReadDir sorts by name, so the first file that is not a test wins, and the first test only when there is nothing else
		if !strings.HasSuffix(name, "_test.go") {
			first = name
			break
		}
		if first == "" {
			first = name
		}
	}

	for _, file := range pass.Files {
		if pass.Fset.Position(file.Pos()).Filename == filepath.Join(dir, first) {
			return file
		}
	}

	return nil
}

// dataFiles returns the files under testdata that a test could be expected to name. ok is false when the folder holds Go files: that is a tree of
// source for a tool to read, such as a linter's own fixtures, and nothing names those one by one.
func dataFiles(testdata string) (files []string, ok bool) {
	ok = true
	_ = filepath.WalkDir(testdata, func(path string, entry fs.DirEntry, err error) error {
		// a folder that is not there or cannot be read has nothing to report
		if err != nil {
			ok = false
			return filepath.SkipAll
		}

		switch {
		case entry.IsDir():
		case strings.HasSuffix(path, ".go"):
			ok = false
			return filepath.SkipAll
		case !strings.HasPrefix(entry.Name(), "."):
			files = append(files, path)
		}

		return nil
	})

	return files, ok
}

// mentions returns every testdata path named by the Go files in dir and beneath it, tests or not, whatever this run is building.
func mentions(dir string) []testdatarefs.Ref {
	var refs []testdatarefs.Ref
	fset := token.NewFileSet()
	_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			name := entry.Name()
			if path != dir && (name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil
		}

		refs = append(refs, testdatarefs.InFile(fset, file)...)

		return nil
	})

	return refs
}
