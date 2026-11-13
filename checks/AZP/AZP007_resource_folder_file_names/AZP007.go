// Package AZP007 defines an analyzer that reports files in a resource or action folder named outside the set those folders use.
package AZP007

import (
	"flag"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks the names of the files in a resource's or action's own folder. Every such folder uses the same few names, so the create of any
// resource is in r_create.go; a file named anything else is somewhere nobody looks.
var Analyzer = &analysis.Analyzer{
	Name: "AZP007",
	Doc:  "check that the files in a resource or action folder use the agreed names",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP007_resource_folder_file_names/README.md",
	Run:  run,
}

// the names a resource or action folder uses
var names = []string{"resource.go", "data_source.go", "list.go", "action.go", "ephemeral.go", "r_create.go", "r_read.go", "r_update.go", "r_delete.go", "r_schema.go", "r_models.go", "r_expand.go", "r_flatten.go"}

// allow adds names to the set, for a provider that has agreed on more.
var allow []string

// internal/services/<service>/resources/<name> or .../actions/<name>, and nothing deeper
var folder = regexp.MustCompile(`/internal/services/[^/]+/(resources|actions)/[^/]+$`)

func init() {
	Analyzer.Flags.Init("AZP007", flag.ContinueOnError)
	Analyzer.Flags.Func("allow", "file name to accept as well; repeat or comma-separate for several", func(s string) error {
		for name := range strings.SplitSeq(s, ",") {
			allow = append(allow, strings.TrimSpace(name))
		}

		return nil
	})
}

func run(pass *analysis.Pass) (any, error) {
	// the main package go test writes to drive a folder's tests has the folder's path, and lives in the build cache
	if pass.Pkg.Name() == "main" || !folder.MatchString(strings.TrimSuffix(pass.Pkg.Path(), "_test")) {
		return nil, nil
	}

	for _, file := range pass.Files {
		name := filepath.Base(pass.Fset.Position(file.Pos()).Filename)

		// tests are named after the file they test, and generated files are named by what generates them
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_gen.go") || slices.Contains(names, name) || slices.Contains(allow, name) {
			continue
		}

		pass.Reportf(file.Name.Pos(), "%s is not one of the file names a resource or action folder uses: move its code into one that is, or into the service's helpers package if it is shared", name)
	}

	return nil, nil
}
