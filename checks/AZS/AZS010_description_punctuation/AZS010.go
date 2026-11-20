// Package AZS010 defines an analyzer that reports punctuation mistakes in resource and property descriptions.
package AZS010

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks how descriptions are punctuated: the capital they start with, the full stop they end with, and the spaces and marks in between.
// The descriptions become the provider's documentation, so a slip here is a slip on the page.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS010",
	Doc:      "check that resource and property descriptions are punctuated properly",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS010_description_punctuation/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// enabled holds, by name, whether each kind of mistake is looked for. All are, until one is switched off.
var enabled = map[string]*bool{}

func init() {
	for _, kind := range [][2]string{
		{"capital", "a description that starts with a lowercase word"},
		{"full-stop", "a description that does not end with a full stop or a question mark"},
		{"question-mark", "a question that ends with a full stop"},
		{"sentence-capital", "a sentence inside a description that starts with a lowercase word"},
		{"leading-mark", "a description that starts with a dash, bullet or other mark"},
		{"edge-space", "a space at the start or the end of a description"},
		{"double-space", "more than one space in a row"},
		{"line-break", "a tab or line break inside a description"},
		{"space-before", "a space before a comma, full stop or other mark"},
		{"space-after", "a comma, full stop or other mark with no space after it"},
		{"bracket-space", "a space just inside a bracket"},
		{"code-space", "code in backticks that touches the word beside it, or has a space just inside the backticks"},
		{"doubled-punctuation", "two marks in a row, such as .. or ,."},
		{"unbalanced", "a backtick, quote or bracket with no partner"},
		{"abbreviation", "e.g., i.e. and etc. written without their full stops"},
		{"pasted-characters", "curly quotes, odd spaces and dashes that come from pasting"},
	} {
		enabled[kind[0]] = Analyzer.Flags.Bool(kind[0], true, "report "+kind[1]+" (false leaves it alone)")
	}
}

var (
	plainWord    = regexp.MustCompile(`^[a-z]+(?:['-][a-z]+)*$`)
	question     = regexp.MustCompile(`^(?:\(\w+\) )?(?:Should|Is|Are|Was|Were|Does|Do|Can|Must|Has|Have|Will) (?:(?:only|also|always|never|either|still) )*([^\s,.;:?!]+)`)
	leadingMark  = regexp.MustCompile(`^[-*\x{2022},;:.!?)\]]+[ \t\n\r]+`)
	spaceBefore  = regexp.MustCompile(`[ \t]+([,.;:?!])`)
	spaceAfter   = regexp.MustCompile(`[,;][A-Za-z\x01]|[A-Za-z]:\x01|(?:\b[a-z]+|\x01|\))[.?!][A-Z][a-z]+(?:[\s,.;:?!)]|$)|[a-z][.?!](?:If|The|This|These|It|When|A|An|In|For|To|Defaults|Changing|Possible|Must|Can|Should|Note|See|Use|Only|Each|One)\s`)
	bracketSpace = regexp.MustCompile(`[(\[][ \t]+|[ \t]+[)\]]`)
	doubleSpace  = regexp.MustCompile(`[ \t]{2,}`)
	lineBreak    = regexp.MustCompile(`[ \t]*[\n\r\t][ \t\n\r]*`)
	codeSpan     = regexp.MustCompile("`[^`]*`")
	marks        = regexp.MustCompile(`[,.;:?!]{2,}`)
	listMarker   = regexp.MustCompile(`(?:^|\s)[A-Za-z0-9]{1,2}$`)
	entity       = regexp.MustCompile(`&#?\w+$`)
)

// the word after "Is", "Must" or "Can" that makes a rule of it, not a question: "Must be between 1 and 5.", "Is required when..."
var statements = strings.Fields("be been not no to have contain include match start begin end one required optional mandatory used ignored set enabled disabled allowed supported deprecated computed returned specified defined expected populated generated limited needed applied configured provided unique valid mutually")

// short forms that end in a full stop without ending the sentence
var shortForms = []string{"e.g.", "i.e.", "etc.", "vs.", "cf.", "approx.", "incl.", "excl.", "viz.", "no.", "min.", "max.", "a.m.", "p.m."}

const (
	curlyQuote   = "a curly quote"
	invisible    = "an invisible character"
	removeSpace  = "Remove the space"
	addSpace     = "Add a space"
	asksQuestion = "description asks a question, so it should end with a question mark"
)

// what pasting from a document leaves behind, and the plain character to use; an empty replacement removes it
var pasted = map[rune][2]string{
	'\u2018': {curlyQuote, "'"},
	'\u2019': {curlyQuote, "'"},
	'\u201c': {curlyQuote, `"`},
	'\u201d': {curlyQuote, `"`},
	'\u2026': {"a single-character ellipsis", "..."},
	'\u2013': {"an en dash", "-"},
	'\u2014': {"an em dash", "-"},
	'\u00a0': {"a non-breaking space", " "},
	'\u202f': {"a non-breaking space", " "},
	'\u2009': {"a thin space", " "},
	'\u200b': {invisible, ""},
	'\u200c': {invisible, ""},
	'\u200d': {invisible, ""},
	'\ufeff': {invisible, ""},
	'\u00ad': {invisible, ""},
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

// check returns what is wrong with a description's punctuation, in the order it appears. A blank description has none to check.
func check(text string) []description.Finding {
	masked := description.Mask(text)
	lo := len(text) - len(strings.TrimLeft(text, " \t\n\r"))
	hi := len(strings.TrimRight(text, " \t\n\r"))
	if lo >= hi {
		return nil
	}

	// a later finding that overlaps an earlier one is dropped, so no two fixes rewrite the same bytes
	var findings []description.Finding
	add := func(kind string, finding description.Finding) {
		if !*enabled[kind] {
			return
		}

		for _, f := range findings {
			if finding.Start < f.End && f.Start < finding.End {
				return
			}
		}
		findings = append(findings, finding)
	}

	if lo > 0 {
		add("edge-space", description.Finding{End: lo, Message: "description should not start with a space", Fix: removeSpace})
	}
	if hi < len(text) {
		add("edge-space", description.Finding{Start: hi, End: len(text), Message: "description should not end with a space", Fix: removeSpace})
	}

	// a dash or a bullet left over from a list; the description proper starts at the word after it
	from := 0
	if at := leadingMark.FindStringIndex(masked[lo:hi]); at != nil {
		from = 1
		add("leading-mark", description.Finding{Start: lo, End: lo + at[1], Message: fmt.Sprintf("description starts with %q - start with its first word", strings.TrimSpace(masked[lo:lo+at[1]])), Fix: "Remove it"})
	}

	for at := 0; at < len(text); {
		r, size := utf8.DecodeRuneInString(text[at:])
		if found, ok := pasted[r]; ok {
			finding := description.Finding{Start: at, End: at + size, Message: fmt.Sprintf("description has %s %q - use %q", found[0], r, found[1]), Fix: "Use the plain character", New: found[1]}
			if found[1] == "" {
				finding.Message = fmt.Sprintf("description has %s (%U) - remove it", found[0], r)
			}

			// an em dash between two words with no spaces would turn them into one hyphenated word
			if r == '\u2014' && (at == 0 || text[at-1] != ' ' || at+size >= len(text) || text[at+size] != ' ') {
				finding.Fix = ""
			}
			add("pasted-characters", finding)
		}
		at += size
	}

	words := description.Words(masked)
	endsAbbreviated := false
	for _, word := range words {
		w := word.Trim(",;:)]}\"'*_")
		var want string
		switch w.Text {
		case "eg", "eg.", "e.g":
			want = "e.g."
		case "Eg", "Eg.", "E.g":
			want = "E.g."
		case "ie", "ie.", "i.e":
			want = "i.e."
		case "Ie", "Ie.", "I.e":
			want = "I.e."
		case "etc":
			want = "etc."
		default:
			continue
		}

		add("abbreviation", description.Finding{Start: w.Start, End: w.End, Message: fmt.Sprintf("%q should be %q", w.Text, want), Fix: "Write " + want, New: want})
		endsAbbreviated = endsAbbreviated || (*enabled["abbreviation"] && w.End == hi)
	}

	for _, at := range spaceBefore.FindAllStringSubmatchIndex(masked[:hi], -1) {
		mark, after := masked[at[2]], byte(' ')
		if at[3] < hi {
			after = masked[at[3]]
		}

		// " .Net" and " :latest" are not a sentence's punctuation, and nor is "..."
		if at[0] < lo || (mark != ',' && !strings.ContainsRune(" \t\n\r)]\"'", rune(after))) {
			continue
		}

		finding := description.Finding{Start: at[0], End: at[2], Message: fmt.Sprintf("description has a space before %q", mark), Fix: removeSpace}
		if at[3] == hi && strings.ContainsRune(",;:", rune(mark)) {
			finding = description.Finding{Start: at[2], End: at[3], Message: fmt.Sprintf("description ends with a %q on its own - if it is a value, put it in backticks", mark)}
		}
		add("space-before", finding)
	}

	for _, at := range bracketSpace.FindAllStringIndex(masked[lo:hi], -1) {
		start, end := lo+at[0], lo+at[1]
		if strings.ContainsRune("([", rune(masked[start])) {
			start++
		} else {
			end--
		}
		add("bracket-space", description.Finding{Start: start, End: end, Message: "description has a space just inside a bracket", Fix: removeSpace})
	}

	for _, at := range lineBreak.FindAllStringIndex(masked[lo:hi], -1) {
		what := "a line break"
		if !strings.ContainsAny(masked[lo+at[0]:lo+at[1]], "\n\r") {
			what = "a tab"
		}
		add("line-break", description.Finding{Start: lo + at[0], End: lo + at[1], Message: "description has " + what + " in it - use a space", Fix: "Use a space", New: " "})
	}

	for _, at := range doubleSpace.FindAllStringIndex(masked[lo:hi], -1) {
		add("double-space", description.Finding{Start: lo + at[0], End: lo + at[1], Message: "description has more than one space in a row", Fix: "Leave one space", New: " "})
	}

	for _, at := range spaceAfter.FindAllStringIndex(masked[lo:hi], -1) {
		at[0] += lo

		// the mark is the first one in the match: a comma, semicolon or colon leads it, a full stop follows a word
		mark := at[0] + strings.IndexAny(masked[at[0]:], ",;:.?!")
		if masked[mark] == ';' && entity.MatchString(masked[:mark]) {
			continue
		}

		// a colon is only short of a space when code follows it: "10:30" and "host:port" are whole
		if masked[mark] == ':' && text[mark+1] != '`' {
			continue
		}
		add("space-after", description.Finding{Start: mark + 1, End: mark + 1, Message: fmt.Sprintf("description has no space after %q", masked[mark]), Fix: addSpace, New: " "})
	}

	// with a backtick missing there is no telling which text is the code
	if strings.Count(text, "`")%2 == 0 {
		for _, at := range codeSpan.FindAllStringIndex(text, -1) {
			start, end := at[0], at[1]
			inner := text[start+1 : end-1]
			if strings.TrimSpace(inner) == "" {
				continue
			}

			if rest := strings.TrimLeft(inner, " "); rest != inner {
				add("code-space", description.Finding{Start: start + 1, End: end - 1 - len(rest), Message: "description has a space just inside a backtick", Fix: removeSpace})
			}
			if rest := strings.TrimRight(inner, " "); rest != inner {
				add("code-space", description.Finding{Start: start + 1 + len(rest), End: end - 1, Message: "description has a space just inside a backtick", Fix: removeSpace})
			}

			if start > 0 && (unicode.IsLetter(rune(text[start-1])) || unicode.IsDigit(rune(text[start-1]))) {
				add("code-space", description.Finding{Start: start, End: start, Message: "description has no space before the code", Fix: addSpace, New: " "})
			}

			// "`subnet`s" is a plural, "`1`and" is a missing space
			if end+1 < len(text) && unicode.IsLower(rune(text[end])) && unicode.IsLower(rune(text[end+1])) {
				add("code-space", description.Finding{Start: end, End: end, Message: "description has no space after the code", Fix: addSpace, New: " "})
			}
			if end < len(text) && text[end] == '`' {
				add("code-space", description.Finding{Start: end - 1, End: end + 1, Message: "description has two pieces of code with nothing between them"})
			}
		}
	}

	for _, at := range marks.FindAllStringIndex(masked[lo:hi], -1) {
		start, end := lo+at[0], lo+at[1]
		found := masked[start:end]

		// "..." trails off, "::" is part of an address, and "e.g.," is a short form followed by a comma
		if found == "..." || strings.Trim(found, ":") == "" {
			continue
		}
		if len(found) == 2 && found[0] == '.' && strings.ContainsRune(",;:", rune(found[1])) && slices.ContainsFunc(shortForms, func(short string) bool { return strings.HasSuffix(strings.ToLower(masked[:start+1]), short) }) {
			continue
		}

		finding := description.Finding{Start: start, End: end, Message: fmt.Sprintf("description has doubled punctuation %q", found)}
		if len(found) == 2 && (found[0] == found[1] || found == "?.") {
			finding.Fix, finding.New = "Leave one mark", found[:1]
		}
		add("doubled-punctuation", finding)
	}

	if strings.Count(text, "`")%2 == 1 {
		at := strings.LastIndex(text, "`")
		add("unbalanced", description.Finding{Start: at, End: at + 1, Message: "description has a backtick with no partner"})
	}
	if strings.Count(masked, `"`)%2 == 1 {
		at := strings.LastIndex(masked, `"`)
		add("unbalanced", description.Finding{Start: at, End: at + 1, Message: "description has a quote with no partner"})
	}

	var open []int
	for at := range len(masked) {
		switch c := masked[at]; c {
		case '(', '[':
			open = append(open, at)
		case ')', ']':
			if len(open) > 0 && (c == ')') == (masked[open[len(open)-1]] == '(') {
				open = open[:len(open)-1]
				continue
			}

			// "a)" and "1)" number the points of a list
			if c == ')' && len(open) == 0 && listMarker.MatchString(masked[:at]) {
				continue
			}
			add("unbalanced", description.Finding{Start: at, End: at + 1, Message: fmt.Sprintf("description has a %q that nothing opened", c)})
		}
	}
	for _, at := range open {
		add("unbalanced", description.Finding{Start: at, End: at + 1, Message: fmt.Sprintf("description has a %q that is never closed", masked[at])})
	}

	first := words[from].Trim(".,;:?!)]}\"'*_")
	if plainWord.MatchString(first.Text) {
		add("capital", description.Finding{Start: first.Start, End: first.Start + 1, Message: "description should start with a capital letter", Fix: "Capitalise it", New: strings.ToUpper(first.Text[:1])})
	}

	// "Should the Thing..." and "Is this..." ask; "Must be..." and "Is required when..." leave the subject out and state a rule
	asks := false
	if opener := question.FindStringSubmatch(masked[lo:hi]); opener != nil {
		asks = !slices.Contains(statements, opener[1])
	}

	sentences := 1
	for i, word := range words[from+1:] {
		before := strings.TrimRight(words[from+i].Text, ")]}\"'*_")
		if before == "" || !strings.ContainsRune(".?!", rune(before[len(before)-1])) || strings.HasSuffix(before, "..") {
			continue
		}

		// a short form, an initial and the number of a list item end in a full stop without ending the sentence
		stem := strings.TrimLeft(strings.ToLower(before), "([{\"'*_")
		if slices.Contains(shortForms, stem) || len(stem) == 2 || (len(stem) > 2 && strings.Trim(stem[:len(stem)-1], "0123456789") == "") {
			continue
		}

		// the question is the first sentence, whatever follows it
		if at := words[from+i].Start + len(before) - 1; sentences == 1 && asks && masked[at] == '.' {
			add("question-mark", description.Finding{Start: at, End: at + 1, Message: asksQuestion, Fix: "End with a question mark", New: "?"})
		}

		sentences++
		if w := word.Trim(".,;:?!)]}\"'*_"); plainWord.MatchString(w.Text) {
			add("sentence-capital", description.Finding{Start: w.Start, End: w.Start + 1, Message: "sentence should start with a capital letter", Fix: "Capitalise it", New: strings.ToUpper(w.Text[:1])})
		}
	}

	// what the description ends with, ignoring a bracket or quote closing it: "(in seconds.)" ends with a full stop
	end := len(strings.TrimRight(masked[:hi], ")]}\"'*_"))
	asks = asks && sentences == 1
	switch {
	case end == 0 || endsAbbreviated || strings.HasSuffix(masked[:end], "\u2026"):
	case masked[end-1] == '?':
	case masked[end-1] == '.' && asks:
		add("question-mark", description.Finding{Start: end - 1, End: end, Message: asksQuestion, Fix: "End with a question mark", New: "?"})
	case masked[end-1] == '.':
	default:
		finding := description.Finding{Start: hi, End: hi, Message: "description should end with a full stop", Fix: "Add a full stop", New: "."}
		if asks {
			finding.Message, finding.Fix, finding.New = asksQuestion, "Add a question mark", "?"
		}

		// after a comma or a colon something is missing, and a full stop would not supply it
		if last, _ := utf8.DecodeLastRuneInString(masked[:hi]); last != description.Blank && !unicode.IsLetter(last) && !unicode.IsDigit(last) && !strings.ContainsRune(")]}\"'*_%", last) {
			finding.Fix = ""
		}
		add("full-stop", finding)
	}

	slices.SortStableFunc(findings, func(a, b description.Finding) int { return cmp.Compare(a.Start, b.Start) })

	return findings
}
