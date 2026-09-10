// Package AZP001 defines an analyzer that reports Microsoft Learn and Docs URLs carrying a
// locale segment (`learn.microsoft.com/en-us/...`) and fixes them by dropping it, so readers
// are served their own language.
package AZP001

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer scans every comment and string literal for a learn, docs, or msdn microsoft.com
// URL whose first path segment is a locale (`/en-us/`, `/en-gb/`) and offers to delete the
// segment; hashicorp/terraform-provider-azurerm#33403 stripped them across the provider.
var Analyzer = &analysis.Analyzer{
	Name:     "AZP001",
	Doc:      "check for Microsoft docs URLs with a locale segment (/en-us/), which should be dropped so readers get their own language",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP001_microsoft_docs_url_locale/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// hosts are the documentation hosts whose first path segment may be a locale; the scheme is
// left to the text so the host cannot be part of a longer name.
var hosts = []string{"://learn.microsoft.com/", "://docs.microsoft.com/", "://msdn.microsoft.com/"}

// locale matches a locale segment, trailing slash included, at the start of a path.
var locale = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}/`)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, c := range group.List {
				check(pass, c.Pos(), c.Text)
			}
		}
	}
	insp.Preorder([]ast.Node{(*ast.BasicLit)(nil)}, func(n ast.Node) {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			check(pass, lit.Pos(), lit.Value)
		}
	})

	return nil, nil
}

// check reports every locale segment in text, whose first byte sits at base in the source.
func check(pass *analysis.Pass, base token.Pos, text string) {
	for _, host := range hosts {
		for from := 0; ; {
			i := strings.Index(text[from:], host)
			if i < 0 {
				break
			}
			at := from + i + len(host)
			if m := locale.FindStringIndex(text[at:]); m != nil {
				start, end := base+token.Pos(at), base+token.Pos(at+m[1])
				pass.Report(analysis.Diagnostic{
					Pos:     base + token.Pos(from+i+len("://")),
					End:     end,
					Message: "Microsoft docs URL has a locale segment `" + text[at:at+m[1]] + "` - drop it so readers get their own language",
					SuggestedFixes: []analysis.SuggestedFix{{
						Message:   "Drop the locale segment",
						TextEdits: []analysis.TextEdit{{Pos: start, End: end}},
					}},
				})
			}
			from = at
		}
	}
}
