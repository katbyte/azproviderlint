// Package AZS018 defines an analyzer that reports a note written into a resource or property description.
package AZS018

import (
	"flag"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a description has no note in it, such as "**Note:** only valid with a subnet". A note is a callout on the docs page, so
// it belongs there and not in the line that says what the property is.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS018",
	Doc:      "check that resource and property descriptions do not have a note in them",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS018_description_note/README.md",
	Requires: []*analysis.Analyzer{description.Analyzer},
	Run:      run,
}

// labels holds, in lowercase, the words a note opens with.
var labels = map[string]bool{"note": true}

// opens finds where a note starts: its label in bold anywhere, or plain with a colon where a sentence starts.
var opens *regexp.Regexp

func init() {
	Analyzer.Flags.Init("AZS018", flag.ContinueOnError)
	Analyzer.Flags.Func("label", "another word a note opens with, such as Important; repeat for several", func(label string) error {
		labels[strings.ToLower(label)] = true
		relabel()

		return nil
	})
	Analyzer.Flags.Func("ignore", "a word a note opens with to leave alone; repeat for several", func(label string) error {
		delete(labels, strings.ToLower(label))
		relabel()

		return nil
	})

	relabel()
}

// relabel rebuilds opens from the labels, after one is added or taken away.
func relabel() {
	bold, plain := make([]string, 0, len(labels)), make([]string, 0, 2*len(labels))
	for _, label := range slices.Sorted(maps.Keys(labels)) {
		bold = append(bold, regexp.QuoteMeta(label))
		plain = append(plain, regexp.QuoteMeta(strings.ToUpper(label[:1])+label[1:]), regexp.QuoteMeta(strings.ToUpper(label)))
	}

	// the docs mark a note with an arrow, which goes with it
	const arrow = `(?:[~!-]> ?)?`
	opens = regexp.MustCompile(`(` + arrow + `(?:\*\*|__) ?(?i:` + strings.Join(bold, "|") + `) ?:? ?(?:\*\*|__) ?:?)|(?:^|[.?!]\)? +)(` + arrow + `(?:` + strings.Join(plain, "|") + `) ?:)`)
}

func run(pass *analysis.Pass) (any, error) {
	for _, found := range description.All(pass) {
		for _, finding := range check(found.Text) {
			found.Report(pass, finding)
		}
	}

	return nil, nil
}

// check returns the note in a description, from its label to the end. It is taken out with the space before it, unless nothing would be left.
func check(text string) []description.Finding {
	if len(labels) == 0 {
		return nil
	}

	at := opens.FindStringSubmatchIndex(description.Mask(text))
	if at == nil {
		return nil
	}

	start := max(at[2], at[4])
	finding := description.Finding{Start: start, End: len(text), Message: "description has a note in it - a note belongs in the docs, not the description"}
	if kept := len(strings.TrimRight(text[:start], " \t\n\r")); kept > 0 {
		finding.Start, finding.Fix = kept, "Remove the note"
	}

	return []description.Finding{finding}
}
