// Package AZT009 defines an analyzer that reports a test naming a testdata file that is not there.
package AZT009

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/checks/azignore"
	"github.com/katbyte/azproviderlint/lib/testdatarefs"
)

// Analyzer checks that every testdata file a test names exists. A test runs in its own folder, so `testdata/cert.pfx` means the file next to the
// test, and a test moved to another folder without its test data only fails once it is run.
var Analyzer = &analysis.Analyzer{
	Name: "AZT009",
	Doc:  "check that the testdata files a test names exist",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT009_testdata_file_missing/README.md",
	Run:  run,
}

// how many folders up to look for the file a test meant
const maxUp = 6

func run(pass *analysis.Pass) (any, error) {
	ignored := azignore.Lines(pass, "AZT009")

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if !strings.HasSuffix(filename, "_test.go") {
			continue
		}

		reported := map[string]bool{}
		for _, ref := range testdatarefs.InFile(pass.Fset, file) {
			// the folder alone, a name finished at run time and the end of a longer path all say too little to look up
			if ref.Name == "" || ref.Pattern || !ref.Rooted || reported[ref.Path] {
				continue
			}

			if _, err := os.Stat(ref.Path); !os.IsNotExist(err) {
				continue
			}

			// no comment can go inside a string that runs over many lines, so one on the line the string starts counts as well
			if ignored[filename][pass.Fset.Position(ref.Lit).Line] {
				continue
			}

			reported[ref.Path] = true

			dir := filepath.Dir(filename)
			if found := nearest(dir, ref.Name); found != "" {
				pass.Reportf(ref.Pos, "%s does not exist: a test reads test data from its own folder, and the nearest copy is %s", ref.Text, found)
				continue
			}

			pass.Reportf(ref.Pos, "%s does not exist: a test reads test data from its own folder", ref.Text)
		}
	}

	return nil, nil
}

// nearest returns where a testdata file of that name does exist, as a path from dir: in a folder above first, then in one beside it.
func nearest(dir, name string) string {
	up := ".."
	for range maxUp {
		if exists(filepath.Join(dir, up, "testdata", name)) {
			return filepath.ToSlash(filepath.Join(up, "testdata", name))
		}

		// the module's own folder is as far up as a test's files can be
		if exists(filepath.Join(dir, up, "go.mod")) {
			break
		}

		up = filepath.Join(up, "..")
	}

	beside, _ := filepath.Glob(filepath.Join(dir, "..", "*", "testdata", name))
	for _, path := range beside {
		if rel, err := filepath.Rel(dir, path); err == nil {
			return filepath.ToSlash(rel)
		}
	}

	return ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
