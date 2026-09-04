// Package azignore implements '//azignore:AZX001 - reason' comment directives, letting
// individual checks be suppressed per line without disabling every azproviderlint check on
// that line the way '//nolint:azproviderlint' does. Directives work under any driver
// (golangci-lint plugin, standalone binary or go vet) as filtering happens inside the
// analyzers themselves. Directives are required to carry a reason — bare ones still
// suppress, but are reported by the AZG000 check.
package azignore

import (
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const prefix = "azignore:"

// ruleList matches the directive's leading comma-separated rule names. Anchoring on the
// rule-name shape (letters then digits, e.g. AZR001) is what lets the reason be free text
// with no mandatory separator: the list ends at the first token that is not rule-shaped.
var ruleList = regexp.MustCompile(`^[A-Za-z]+\d+(\s*,\s*[A-Za-z]+\d+)*`)

// ParseDirective splits a comment's text into the azignore directive's rule names and
// reason. ok is false when the comment is not an azignore directive at all. The reason is
// everything after the rule list, less any leading '-'/'–'/'—' separator, which is
// permitted but not required: 'azignore:AZR001 - deliberate' and 'azignore:AZR001
// deliberate' parse the same. A directive without a reason returns ok true with an empty
// reason — Lines still honours it (so a bad suppression never surfaces as a confusing
// report from the suppressed check) and AZG000 reports the directive itself.
func ParseDirective(text string) (rules []string, reason string, ok bool) {
	body := strings.TrimSpace(strings.TrimPrefix(text, "//"))
	if !strings.HasPrefix(body, prefix) {
		return nil, "", false
	}
	body = strings.TrimSpace(strings.TrimPrefix(body, prefix))

	list := ruleList.FindString(body)
	for r := range strings.SplitSeq(list, ",") {
		if r = strings.TrimSpace(r); r != "" {
			rules = append(rules, r)
		}
	}

	reason = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(body[len(list):]), ",-–—"))

	return rules, reason, true
}

// Wrap replaces each analyzer's Run with one that drops diagnostics on lines carrying a
// '//azignore:<Name> - <reason>' comment, either at the end of the line or on the line
// immediately preceding it. Multiple checks can be listed: '//azignore:AZG001,AZR001 - why'.
func Wrap(analyzers []*analysis.Analyzer) []*analysis.Analyzer {
	for _, a := range analyzers {
		wrap(a)
	}
	return analyzers
}

func wrap(a *analysis.Analyzer) {
	run := a.Run
	name := a.Name
	a.Run = func(pass *analysis.Pass) (any, error) {
		ignored := Lines(pass, name)
		report := pass.Report
		pass.Report = func(d analysis.Diagnostic) {
			pos := pass.Fset.Position(d.Pos)
			if lines, ok := ignored[pos.Filename]; ok && lines[pos.Line] {
				return
			}
			report(d)
		}
		return run(pass)
	}
}

// Lines collects, per filename, the lines on which diagnostics from the named analyzer are
// suppressed: the line of each matching directive and the line below it. It is exported so
// checks whose reports aggregate several source positions (e.g. AZS006, which reports every
// missing property on the data source's registration line) can honour directives placed on
// the individual positions too.
func Lines(pass *analysis.Pass, name string) map[string]map[int]bool {
	ignored := map[string]map[int]bool{}

	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, comment := range group.List {
				rules, _, ok := ParseDirective(comment.Text)
				if !ok || !slices.Contains(rules, name) {
					continue
				}

				pos := pass.Fset.Position(comment.Pos())
				if ignored[pos.Filename] == nil {
					ignored[pos.Filename] = map[int]bool{}
				}
				ignored[pos.Filename][pos.Line] = true
				ignored[pos.Filename][pos.Line+1] = true
			}
		}
	}

	return ignored
}

// ExactLines collects, per filename, only the lines carrying a matching directive for the
// named analyzer — unlike Lines, it does not also mark the line below. Checks that scope a
// suppression to an exact line (e.g. a directive on a composite literal's opening line) use
// this so a trailing directive does not leak onto the following line.
func ExactLines(pass *analysis.Pass, name string) map[string]map[int]bool {
	ignored := map[string]map[int]bool{}

	for _, file := range pass.Files {
		for _, group := range file.Comments {
			for _, comment := range group.List {
				rules, _, ok := ParseDirective(comment.Text)
				if !ok || !slices.Contains(rules, name) {
					continue
				}

				pos := pass.Fset.Position(comment.Pos())
				if ignored[pos.Filename] == nil {
					ignored[pos.Filename] = map[int]bool{}
				}
				ignored[pos.Filename][pos.Line] = true
			}
		}
	}

	return ignored
}
