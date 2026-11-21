// Package AZS017 defines an analyzer that reports resource and property descriptions that open with a verb such as "Specifies".
package AZS017

import (
	"flag"
	"fmt"
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that a description says what the property is, "The name of the Thing.", the way most of the provider does, and does not open
// with a verb that only says it is a description: "Specifies the name of the Thing."
var Analyzer = &analysis.Analyzer{
	Name:     "AZS017",
	Doc:      `check that resource and property descriptions do not open with a verb such as "Specifies"`,
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS017_description_opens_with_specifies/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// verbs holds, in lowercase, the opening verbs that add nothing to what follows them.
var verbs = map[string]bool{"specifies": true, "specify": true, "indicates": true, "denotes": true, "defines": true, "determines": true}

// openers holds what a description opens with once the verb is gone, by the word that follows the verb. After any other word there is no fix.
var openers = map[string]string{"the": "The", "a": "A", "an": "An", "whether": "Whether", "if": "Whether", "which": "Which", "how": "How", "where": "Where", "what": "What"}

func init() {
	Analyzer.Flags.Init("AZS017", flag.ContinueOnError)
	Analyzer.Flags.Func("verb", "another opening verb to report; repeat for several", func(verb string) error {
		verbs[strings.ToLower(verb)] = true
		return nil
	})
	Analyzer.Flags.Func("ignore", "an opening verb to leave alone; repeat for several", func(verb string) error {
		delete(verbs, strings.ToLower(verb))
		return nil
	})
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	for _, found := range description.All(pass, insp) {
		for _, finding := range check(found.Text) {
			found.Report(pass, finding)
		}
	}

	return nil, nil
}

// check returns the verb the description opens with, when it is one that adds nothing.
func check(text string) []description.Finding {
	words := description.Words(description.Mask(text))
	if len(words) < 2 || !verbs[strings.ToLower(words[0].Text)] {
		return nil
	}

	verb, next := words[0], words[1].Trim(".,;:?!")
	if opener, ok := openers[strings.ToLower(next.Text)]; ok && next.Start == words[1].Start {
		return []description.Finding{{Start: verb.Start, End: next.End, Message: fmt.Sprintf("description opens with %q - open with %q", verb.Text+" "+next.Text, opener), Fix: "Open with " + opener, New: opener}}
	}

	return []description.Finding{{Start: verb.Start, End: verb.End, Message: fmt.Sprintf("description opens with %q - say what the property is instead", verb.Text)}}
}
