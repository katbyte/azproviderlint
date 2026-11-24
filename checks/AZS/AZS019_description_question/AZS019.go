// Package AZS019 defines an analyzer that reports a resource or property description written as a question.
package AZS019

import (
	"go/ast"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a description says what a property is and does not ask it, so "Should the Thing be enabled?" is reported. The docs
// describe a setting; they do not put a question to the reader.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS019",
	Doc:      "check that resource and property descriptions are not written as a question",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS019_description_question/README.md",
	Requires: []*analysis.Analyzer{description.Analyzer},
	Run:      run,
}

var (
	// the words a question opens with, behind a tag such as "(Boolean)" if there is one, and the word that follows
	opener = regexp.MustCompile(`^(?:\(\w+\) )?(?:Should|Is|Are|Was|Were|Does|Do|Can|Must|Has|Have|Will) (?:(?:only|also|always|never|either|still) )*([^\s,.;:?!]+)`)

	// the parts of a first sentence that can be turned round: a tag such as "(Boolean) ", the word it opens with, the rest, and its closing marks
	question = regexp.MustCompile(`^((?:\(\w+\) )?)(Should|Can|Must|Will|Is|Are|Whether) (.*?)([?.]*)$`)

	// the verb a "Should..." question turns on: the Thing "be" enabled, "use" HTTPS, "have" a Public IP. None of them is also a common noun, as
	// "access" and "support" are.
	verbs = regexp.MustCompile(` (be|use|have|recover|receive|require|keep|delete|validate|send|generate|retain|skip|restart|operate|expose|allow|disable|inspect|retrieve)\b`)

	// what an "Is..." question turns on, and what may follow it: "enabled", "required to...", "used for..."
	states = regexp.MustCompile(` ((?i:enabled|disabled|required|permitted|allowed|supported|enforced|used|visible|accessible|hidden))(?:$|,| (?:for|on|in|by|to|when|if|at|with|as|or|from|within|via|during|across|through|after|before|inside|over)\b)`)

	// "automatically" in "Should the Thing automatically be replaced", which stays next to the verb
	adverb = regexp.MustCompile(` (?:\w+ly|only|also|always|never)$`)

	// what shows a "Should..." question is garbled: "Should the Thing that is enabled for the App be enabled"
	otherVerb = regexp.MustCompile(`\b(?:is|are|was|were)\b`)

	// where the noun a subject hangs off ends, and the nouns ending in "s" that are one thing
	hangsOff = regexp.MustCompile(` (?:of|for|in|on|to|from|with|that|which|within|[a-z]+ed (?:in|on|to|by|for|with|from|at)) | \(`)
	oneThing = regexp.MustCompile(`(?:ss|us|is)$|^(?:Kubernetes|Redis|Analytics|Insights|Windows)$`)
	several  = regexp.MustCompile(` and |^(?:we|you|they)$`)
)

// the word after "Is", "Must" or "Can" that makes a rule of it, not a question: "Must be between 1 and 5.", "Is required when..."
var statements = strings.Fields("be been not no to have contain include match start begin end one required optional mandatory used ignored set enabled disabled allowed supported deprecated computed returned specified defined expected populated generated limited needed applied configured provided unique valid mutually")

func run(pass *analysis.Pass) (any, error) {
	for _, found := range description.All(pass) {
		readOnly := false
		if found.Owner != nil {
			fields := map[string]ast.Expr{}
			for _, elt := range found.Owner.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if key, ok := kv.Key.(*ast.Ident); ok {
						fields[key.Name] = kv.Value
					}
				}
			}
			readOnly = astx.IsTrueConstant(pass, fields["Computed"]) && fields["Optional"] == nil && fields["Required"] == nil
		}

		for _, finding := range check(found.Text, readOnly) {
			found.Report(pass, finding)
		}
	}

	return nil, nil
}

// a word a subject cannot end on, which shows the verb found is not the question's own
var notSubject = strings.Fields("the a an this that these those its their any all each every no of for to in on with and or not be is are was were which who could can cannot will would must may might shall should")

// check returns each question in a description: a first sentence that opens like one, and any sentence ending in a question mark. A read-only
// property's question is turned into what it is, not what it should be.
func check(text string, readOnly bool) []description.Finding {
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

		if parts := question.FindStringSubmatch(text[sentence.Start:sentence.End]); i == 0 && parts != nil {
			rest, verb, statement := parts[3], strings.ToLower(parts[2]), ""

			// "Is the Thing enabled" turns on a state, "Should the Thing use HTTPS" on a verb; "Whether..." only ends wrongly
			turns, state := [][]int{}, false
			switch verb {
			case "is", "are":
				turns, state = states.FindAllStringSubmatchIndex(rest, -1), true
			case "whether":
				statement = rest
			default:
				turns = verbs.FindAllStringSubmatchIndex(rest, -1)
			}

			// the subject ends where the question's own verb or state begins, which is the first one a subject can end in front of
			for _, at := range turns {
				subject, turn := rest[:at[2]-1], rest[at[2]:at[3]]

				// with a "be" in it the question turns on that: "Should the use of HTTPS be required"; a verb found after a comma is not the question's
				if !state && turn != "be" && (strings.Contains(rest, " be ") || strings.Contains(subject, ",")) {
					continue
				}

				if subject == "" || strings.HasPrefix(subject, "there ") || slices.Contains(notSubject, strings.ToLower(subject[strings.LastIndex(subject, " ")+1:])) {
					continue
				}

				// a second "should" or another verb means the question is garbled, and moving a word would not mend it
				if !state && (strings.Contains(rest, " "+verb+" ") || otherVerb.MatchString(subject)) {
					break
				}

				how := adverb.FindString(subject)
				subject = strings.TrimSuffix(subject, how)

				// nothing is asked of a read-only property, so it "is enabled" or "uses HTTPS" where one that is set "should be enabled" or "should
				// use HTTPS"; "are" and "use" when the noun the subject hangs off is plural
				said := verb + how + " " + turn
				if readOnly && verb == "should" {
					noun := hangsOff.Split(subject, 2)[0]
					last := noun[strings.LastIndex(noun, " ")+1:]
					plural := several.MatchString(noun) || (strings.HasSuffix(last, "s") && last != strings.ToUpper(last) && !strings.Contains(last, "`") && !oneThing.MatchString(last))

					switch {
					case turn == "be" && plural:
						said = "are" + how
					case turn == "be":
						said = "is" + how
					case plural:
						said = how + " " + turn
					case turn == "have":
						said = how + " has"
					default:
						said = how + " " + turn + "s"
					}
				}

				statement = subject + " " + strings.TrimSpace(said) + rest[at[3]:]

				break
			}

			if statement != "" {
				finding.Fix, finding.New = "Write it as a statement", parts[1]+"Whether "+statement
				if parts[4] != "" {
					finding.New += "."
				}
			}
		}
		findings = append(findings, finding)
	}

	return findings
}
