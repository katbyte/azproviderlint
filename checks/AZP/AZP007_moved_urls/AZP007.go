// Package AZP007 defines an analyzer that reports URLs whose page has moved, and fixes them to the new address.
package AZP007

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks every URL in a comment or a string against the addresses known to have moved: the Microsoft docs to learn.microsoft.com, the
// Terraform docs to the registry and developer.hashicorp.com. An old address works only for as long as its redirect is kept up.
var Analyzer = &analysis.Analyzer{
	Name:     "AZP007",
	Doc:      "check for URLs whose page has moved, such as docs.microsoft.com to learn.microsoft.com",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP007_moved_urls/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// extra is more moves for the consumer to add, as old=new pairs of what the address starts with.
var extra string

func init() {
	Analyzer.Flags.StringVar(&extra, "moved", "", "comma-separated old=new pairs of what a moved URL starts with, after the scheme: old.example.com/docs/=new.example.com/")
}

// move is an address that has moved: what it looks like after the scheme, and what to write in its place. With nothing to write, the page has
// moved but not to somewhere that can be worked out from the old address.
type move struct {
	from *regexp.Regexp
	to   string
}

// the first one to match an address is the one used, so the more exact come first
var moves = []move{
	{regexp.MustCompile(`^docs\.microsoft\.com/`), "learn.microsoft.com/"},
	{regexp.MustCompile(`^(?:www\.)?terraform\.io/docs/providers/(azurerm|azuread|azurestack)/r/([a-z0-9_]+)\.html`), "registry.terraform.io/providers/hashicorp/$1/latest/docs/resources/$2"},
	{regexp.MustCompile(`^(?:www\.)?terraform\.io/docs/providers/(azurerm|azuread|azurestack)/d/([a-z0-9_]+)\.html`), "registry.terraform.io/providers/hashicorp/$1/latest/docs/data-sources/$2"},
	{regexp.MustCompile(`^(?:www\.)?terraform\.io/docs/configuration/functions/([a-z0-9_]+)\.html`), "developer.hashicorp.com/terraform/language/functions/$1"},
	{regexp.MustCompile(`^github\.com/terraform-providers/terraform-provider-(azurerm|azuread|azurestack)\b`), "github.com/hashicorp/terraform-provider-$1"},
	{regexp.MustCompile(`^(?:www\.)?terraform\.io/(?:docs|language)/`), ""},
}

var scheme = regexp.MustCompile(`https?://`)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// the consumer's moves are tried first, so one can stand in for a move built in here
	known := moves
	for pair := range strings.SplitSeq(extra, ",") {
		if from, to, ok := strings.Cut(strings.TrimSpace(pair), "="); ok && from != "" && to != "" {
			known = append([]move{{regexp.MustCompile("^" + regexp.QuoteMeta(from)), strings.ReplaceAll(to, "$", "$$")}}, known...)
		}
	}

	check := func(base token.Pos, text string) {
		for _, at := range scheme.FindAllStringIndex(text, -1) {
			for _, m := range known {
				found := m.from.FindStringSubmatchIndex(text[at[1]:])
				if found == nil {
					continue
				}

				old := text[at[1] : at[1]+found[1]]
				start, end := base+token.Pos(at[1]), base+token.Pos(at[1]+found[1])
				if m.to == "" {
					pass.Report(analysis.Diagnostic{Pos: start, End: end, Message: "URL starts with `" + old + "`, a page that has moved - look up where it lives now"})

					break
				}

				use := string(m.from.ExpandString(nil, m.to, text[at[1]:], found))
				pass.Report(analysis.Diagnostic{Pos: start, End: end, Message: "URL starts with `" + old + "`, which has moved - use `" + use + "`", SuggestedFixes: []analysis.SuggestedFix{{Message: "Use " + use, TextEdits: []analysis.TextEdit{{Pos: start, End: end, NewText: []byte(use)}}}}})

				break
			}
		}
	}

	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, c := range group.List {
				check(c.Pos(), c.Text)
			}
		}
	}

	insp.Preorder([]ast.Node{(*ast.BasicLit)(nil)}, func(n ast.Node) {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			check(lit.Pos(), lit.Value)
		}
	})

	return nil, nil
}
