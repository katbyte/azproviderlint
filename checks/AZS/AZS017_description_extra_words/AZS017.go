// Package AZS017 defines an analyzer that reports words in resource and property descriptions that add nothing, such as "Specifies" and "or not".
package AZS017

import (
	"cmp"
	"flag"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that a description says what the property is in as few words as it takes: "The name of the Thing.", not "Specifies the name of
// the Thing.", and "Whether to cache.", not "Whether to cache or not."
var Analyzer = &analysis.Analyzer{
	Name:     "AZS017",
	Doc:      `check that resource and property descriptions have no extra words, such as "Specifies" and "or not"`,
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS017_description_extra_words/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// verbs holds, in lowercase, the opening verbs that add nothing to what follows them.
var verbs = map[string]bool{"specifies": true, "specify": true, "indicates": true, "denotes": true, "defines": true, "determines": true}

// openers holds what a description opens with once the verb is gone, by the word that follows the verb. After any other word there is no fix.
var openers = map[string]string{"the": "The", "a": "A", "an": "An", "whether": "Whether", "if": "Whether", "which": "Which", "how": "How", "where": "Where", "what": "What"}

// phrases holds, in lowercase, the wordy phrases and the shorter way to say each; one with nothing to say instead is just taken out.
var phrases = map[string]string{
	"or not":                  "",
	"in order to":             "to",
	"in order for":            "for",
	"due to the fact that":    "because",
	"in the event that":       "if",
	"for the purpose of":      "for",
	"is able to":              "can",
	"are able to":             "can",
	"prior to":                "before",
	"please":                  "",
	"please note":             "",
	"please note that":        "",
	"note that":               "",
	"it should be noted that": "",
}

// wordy finds the phrases, the longest first so "please note that" is not read as "please".
var wordy *regexp.Regexp

func init() {
	Analyzer.Flags.Init("AZS017", flag.ContinueOnError)
	Analyzer.Flags.Func("verb", "another opening verb to report; repeat for several", func(verb string) error {
		verbs[strings.ToLower(verb)] = true
		return nil
	})
	Analyzer.Flags.Func("phrase", `another wordy phrase to report and what to say instead, as "in order to=to"; nothing after the = takes it out; repeat for several`, func(phrase string) error {
		long, shorter, _ := strings.Cut(phrase, "=")
		phrases[strings.ToLower(strings.TrimSpace(long))] = strings.TrimSpace(shorter)
		rebuild()

		return nil
	})
	Analyzer.Flags.Func("ignore", "an opening verb or a phrase to leave alone; repeat for several", func(word string) error {
		delete(verbs, strings.ToLower(word))
		delete(phrases, strings.ToLower(word))
		rebuild()

		return nil
	})

	rebuild()
}

// rebuild makes wordy from the phrases, after one is added or taken away.
func rebuild() {
	longest := slices.SortedFunc(maps.Keys(phrases), func(a, b string) int { return cmp.Or(cmp.Compare(len(b), len(a)), cmp.Compare(a, b)) })
	for i, phrase := range longest {
		longest[i] = regexp.QuoteMeta(phrase)
	}

	wordy = regexp.MustCompile(`(?i)\b(?:` + strings.Join(longest, "|") + `)\b`)
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

// check returns the words in a description that add nothing: a verb it opens with that only says it is a description, and wordy phrases.
func check(text string) []description.Finding {
	masked := description.Mask(text)

	var findings []description.Finding
	if words := description.Words(masked); len(words) > 1 && verbs[strings.ToLower(words[0].Text)] {
		verb, next := words[0], words[1].Trim(".,;:?!")
		if opener, ok := openers[strings.ToLower(next.Text)]; ok && next.Start == words[1].Start {
			findings = append(findings, description.Finding{Start: verb.Start, End: next.End, Message: fmt.Sprintf("description opens with %q - open with %q", verb.Text+" "+next.Text, opener), Fix: "Open with " + opener, New: opener})
		} else {
			findings = append(findings, description.Finding{Start: verb.Start, End: verb.End, Message: fmt.Sprintf("description opens with %q - say what the property is instead", verb.Text)})
		}
	}

	if len(phrases) == 0 {
		return findings
	}

	for _, at := range wordy.FindAllStringIndex(masked, -1) {
		start, end := at[0], at[1]
		said, before, after := text[start:end], strings.ToLower(masked[:start]), masked[end:]
		shorter := phrases[strings.ToLower(said)]

		// "or not" is only extra after "whether" or closing a clause: "empty or not set" needs it
		if strings.EqualFold(said, "or not") && !strings.HasSuffix(before, "whether ") && after != "" && !strings.ContainsRune(".,;:?!)", rune(after[0])) {
			continue
		}

		// where a sentence opened with the phrase, what takes its place opens it
		opens := unicode.IsUpper(rune(said[0]))
		finding := description.Finding{Start: start, End: end, Message: fmt.Sprintf("description has %q - it adds nothing", said), Fix: "Remove it"}
		switch {
		case shorter != "":
			if opens {
				shorter = strings.ToUpper(shorter[:1]) + shorter[1:]
			}
			finding.Message, finding.Fix, finding.New = fmt.Sprintf("description has %q - %q says the same", said, shorter), fmt.Sprintf("Use %q", shorter), shorter

			// "agree to in order to sign up" would be left with "to" twice in a row
			if strings.HasSuffix(before, " "+strings.ToLower(shorter)+" ") {
				finding.Fix, finding.New = "", ""
			}

		case opens && len(after) > 1 && after[0] == ' ' && after[1] != ' ':
			finding.End++
			if unicode.IsLower(rune(after[1])) {
				finding.End, finding.New = finding.End+1, strings.ToUpper(after[1:2])
			}

		case !opens && strings.HasSuffix(before, " "):
			finding.Start--

		case !opens && strings.HasPrefix(after, " "):
			finding.End++

		default:
			finding.Fix = ""
		}
		findings = append(findings, finding)
	}

	return findings
}
