// Package AZS019 defines an analyzer that reports a resource or property description written as a question.
package AZS019

import (
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that a description says what a property is and does not ask it, so "Should the Thing be enabled?" is reported. The docs
// describe a setting; they do not put a question to the reader.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS019",
	Doc:      "check that resource and property descriptions are not written as a question",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS019_description_question/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var (
	// the words a question opens with, behind a tag such as "(Boolean)" if there is one, and the word that follows
	opener = regexp.MustCompile(`^(?:\(\w+\) )?(?:Should|Is|Are|Was|Were|Does|Do|Can|Must|Has|Have|Will) (?:(?:only|also|always|never|either|still) )*([^\s,.;:?!]+)`)

	// "Should the Thing be enabled", the one question that turns into a statement by moving a word
	shouldBe = regexp.MustCompile(`^((?:\(\w+\) )?)Should (.+?) be (.+?)([?.]*)$`)
)

// the word after "Is", "Must" or "Can" that makes a rule of it, not a question: "Must be between 1 and 5.", "Is required when..."
var statements = strings.Fields("be been not no to have contain include match start begin end one required optional mandatory used ignored set enabled disabled allowed supported deprecated computed returned specified defined expected populated generated limited needed applied configured provided unique valid mutually")

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

// check returns each question in a description: a first sentence that opens like one, and any sentence ending in a question mark.
func check(text string) []description.Finding {
	var findings []description.Finding
	for i, sentence := range description.Sentences(description.Mask(text)) {
		asks := strings.HasSuffix(strings.TrimRight(sentence.Text, ".)]}\"'*_"), "?")
		if found := opener.FindStringSubmatch(sentence.Text); i == 0 && found != nil && !slices.Contains(statements, found[1]) {
			asks = true
		}
		if !asks {
			continue
		}

		finding := description.Finding{Start: sentence.Start, End: sentence.End, Message: `description asks a question - say what it is, such as "Whether the Thing is enabled."`}

		// a second "should" means the question is garbled, and moving a word would not mend it
		if parts := shouldBe.FindStringSubmatch(text[sentence.Start:sentence.End]); i == 0 && parts != nil && !strings.Contains(parts[2]+" "+parts[3], "should ") {
			finding.Fix, finding.New = "Write it as a statement", parts[1]+"Whether "+parts[2]+" should be "+parts[3]
			if parts[4] != "" {
				finding.New += "."
			}
		}
		findings = append(findings, finding)
	}

	return findings
}
