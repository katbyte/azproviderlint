// Package AZS020 defines an analyzer that reports a resource or property description naming a resource or property the provider does not have.
package AZS020

import (
	"flag"
	"fmt"
	"go/constant"
	"go/scanner"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a name in backticks in a description is one the provider has, so a description still pointing at `client_cert_enabled`
// after the property became `client_certificate_enabled` is reported. A reader who follows a name that is not there finds nothing.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS020",
	Doc:      "check that resource and property descriptions do not name a resource or property that does not exist",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS020_description_unknown_name/README.md",
	Requires: []*analysis.Analyzer{description.Analyzer},
	Run:      run,
}

// allowed holds the names that are fine though the provider does not define them. It starts with Terraform's own.
var allowed = map[string]bool{"for_each": true, "depends_on": true, "ignore_changes": true, "create_before_destroy": true, "prevent_destroy": true, "replace_triggered_by": true}

func init() {
	Analyzer.Flags.Init("AZS020", flag.ContinueOnError)
	Analyzer.Flags.Func("allow", "a name to accept though the provider does not define it, such as another provider's resource; repeat for several", func(name string) error {
		allowed[name] = true

		return nil
	})
}

var (
	code = regexp.MustCompile("`([^`]+)`")

	// a resource's name, at the start of what is in the backticks: "example_thing" of "example_thing.main.id"
	leading = regexp.MustCompile(`^[a-z0-9_]+`)

	// a property, or the way to one: "site_config.0.linux_fx_version", "identity[0].principal_id"
	path = regexp.MustCompile(`^[a-z0-9_.\[\]]+$`)
	part = regexp.MustCompile(`[^.\[\]]+`)

	// one word on its own, such as "enabled", may as well be a value, so it is a name only when it has two words or the text calls it one
	name   = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
	called = regexp.MustCompile(`^ (?:block|property|field|argument|attribute)s?\b`)

	module = regexp.MustCompile(`(?m)^module\s+\S*?([^/\s"]+)"?\s*$`)
	tagged = regexp.MustCompile(`tfschema:"([a-z0-9_]+)"`)
)

// provider is what one provider's code holds: the word its resources start with, and every name written out in it.
type provider struct {
	prefix string
	read   sync.Once
	names  map[string]bool
}

var (
	providers = map[string]*provider{}
	lock      sync.Mutex
)

func run(pass *analysis.Pass) (any, error) {
	found := description.All(pass)
	if len(found) == 0 {
		return nil, nil
	}

	// the provider is everything under the nearest go.mod, which also names it: terraform-provider-example has resources called example_...
	var mod []byte
	root := filepath.Dir(pass.Fset.Position(found[0].Value.Pos()).Filename)
	for {
		var err error
		if mod, err = os.ReadFile(filepath.Join(root, "go.mod")); err == nil { //nolint:gosec // the go.mod of the code being checked
			break
		}
		if root == filepath.Dir(root) {
			return nil, nil
		}
		root = filepath.Dir(root)
	}

	lock.Lock()
	p := providers[root]
	if p == nil {
		p = &provider{names: map[string]bool{}}
		if m := module.FindSubmatch(mod); m != nil {
			p.prefix = strings.TrimPrefix(string(m[1]), "terraform-provider-")
		}
		providers[root] = p
	}
	lock.Unlock()

	known := func(said string) bool {
		if allowed[said] {
			return true
		}

		// the whole provider is read once, the first time a description names something
		p.read.Do(func() {
			_ = filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
				// a folder that cannot be read is left out
				if err != nil {
					return fs.SkipDir
				}

				if entry.IsDir() {
					if dir := entry.Name(); file != root && (dir == "vendor" || dir == "testdata" || strings.HasPrefix(dir, ".") || strings.HasPrefix(dir, "_")) {
						return filepath.SkipDir
					}

					return nil
				}

				if !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
					return nil
				}

				src, _ := os.ReadFile(file) //nolint:gosec // a file of the code being checked

				// every piece of text the code writes out that is a name: a schema key, a resource's type, a value, a model's tfschema tag
				var s scanner.Scanner
				s.Init(token.NewFileSet().AddFile(file, -1, len(src)), src, nil, 0)
				for {
					_, tok, text := s.Scan()
					if tok == token.EOF {
						break
					}
					if tok != token.STRING || len(text) < 2 {
						continue
					}

					for _, tag := range tagged.FindAllStringSubmatch(text, -1) {
						p.names[tag[1]] = true
					}
					if text = text[1 : len(text)-1]; name.MatchString(text) {
						p.names[text] = true
					}
				}

				return nil
			})
		})

		if p.names[said] {
			return true
		}

		// a value may come from outside the provider, as one of an SDK's constants
		for _, imported := range pass.Pkg.Imports() {
			for _, each := range imported.Scope().Names() {
				if c, ok := imported.Scope().Lookup(each).(*types.Const); ok && c.Val().Kind() == constant.String && constant.StringVal(c.Val()) == said {
					return true
				}
			}
		}

		return false
	}

	for _, each := range found {
		for _, finding := range check(each.Text, p.prefix, known) {
			each.Report(pass, finding)
		}
	}

	return nil, nil
}

// check returns each name in backticks that the provider does not have, given the word its resources start with and what tells whether a name
// is known.
func check(text, prefix string, known func(string) bool) []description.Finding {
	var findings []description.Finding
	for _, at := range code.FindAllStringSubmatchIndex(text, -1) {
		said := text[at[2]:at[3]]

		// a resource is named by its type, whatever follows it; "example_thing_*" stands for several and names none
		if prefix != "" && strings.HasPrefix(said, prefix+"_") {
			resource := leading.FindString(said)
			if !strings.HasSuffix(resource, "_") && !strings.HasPrefix(said[len(resource):], "*") && !known(resource) {
				findings = append(findings, description.Finding{Start: at[2], End: at[2] + len(resource), Message: fmt.Sprintf("description names the resource %q - the provider has no resource called that", resource)})
			}

			continue
		}

		if !path.MatchString(said) {
			continue
		}

		for _, seg := range part.FindAllStringIndex(said, -1) {
			if property := said[seg[0]:seg[1]]; name.MatchString(property) && (strings.Contains(property, "_") || called.MatchString(text[at[1]:])) && !known(property) {
				findings = append(findings, description.Finding{Start: at[2] + seg[0], End: at[2] + seg[1], Message: fmt.Sprintf("description names %q - nothing in the provider is called that", property)})
			}
		}
	}

	return findings
}
