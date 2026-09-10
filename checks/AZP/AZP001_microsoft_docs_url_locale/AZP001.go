// Package AZP001 defines an analyzer that reports Microsoft Learn and Docs URLs carrying a
// locale segment (`learn.microsoft.com/en-us/...`) and fixes them by dropping it, so readers
// are served their own language.
package AZP001

import (
	"go/ast"
	"go/token"
	"regexp"

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

// localeURL captures the locale segment, trailing slash included, so deleting the capture
// leaves `host/rest`.
var localeURL = regexp.MustCompile(`https?://(?:learn|docs|msdn)\.microsoft\.com/([a-z]{2}-[a-z]{2}/)`)

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
	for _, m := range localeURL.FindAllStringSubmatchIndex(text, -1) {
		start, end := base+token.Pos(m[2]), base+token.Pos(m[3])
		pass.Report(analysis.Diagnostic{
			Pos:     base + token.Pos(m[0]),
			End:     end,
			Message: "Microsoft docs URL has a locale segment `" + text[m[2]:m[3]] + "` - drop it so readers get their own language",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message:   "Drop the locale segment",
				TextEdits: []analysis.TextEdit{{Pos: start, End: end}},
			}},
		})
	}
}
