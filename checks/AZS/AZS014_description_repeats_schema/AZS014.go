// Package AZS014 defines an analyzer that reports text in resource and property descriptions that repeats what the schema already says.
package AZS014

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks descriptions for what the schema already states: the type, the default, the allowed values, that a change forces a new
// resource. The docs take those from the schema, so a description that repeats them says it twice and goes stale when the schema changes.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS014",
	Doc:      "check that resource and property descriptions do not repeat what the schema says",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS014_description_repeats_schema/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// enabled holds, by name, whether each kind of repeat is looked for. All are, until one is switched off.
var enabled = map[string]*bool{}

func init() {
	for _, kind := range [][2]string{
		{"tag", "a tag copied from the docs, such as (Optional) or (Boolean)"},
		{possible, "a sentence listing the values a property takes, when its validation or its type already does"},
		{"default", "a sentence giving the default, when the property sets one"},
		{"force-new", "a sentence saying a change forces a new resource, when the property is ForceNew"},
		{"conflicts-with", "a sentence saying what a property conflicts with, when it sets ConflictsWith"},
		{"one-of", "a sentence saying one of several properties must be set, when it sets ExactlyOneOf or AtLeastOneOf"},
		{"item-count", "a sentence saying how many items or blocks there can be, when it sets MaxItems or MinItems"},
		{length, "words giving how long a value can be, when its validation already does"},
		{below, "words that point at the docs page, such as \"as defined below\""},
	} {
		enabled[kind[0]] = Analyzer.Flags.Bool(kind[0], true, "report "+kind[1]+" (false leaves it alone)")
	}
}

const (
	possible = "possible-values"
	length   = "length"
	below    = "defined-below"

	givesDefault = "description gives the default - the schema's Default already does"
	givesLength  = "description says how long it can be - the schema's validation already does"

	removeSentence = "Remove the sentence"
	removeWords    = "Remove the words"
)

// what a value, a list of them and a count look like once the code in backticks is blanked out
const (
	value  = `(?:\x01+|'[^']*'|"[^"]*"|[\w-]+)`
	values = value + `(?:(?:, and |, or |, | and | or )` + value + `)+`
	count  = `(?:\x01+|\d+|one|two|three|four|five|six|seven|eight|nine|ten)`
	span   = `(?:between ` + count + ` and ` + count + `|\d+-\d+|up to ` + count + `|at most ` + count + `|at least ` + count + `|a maximum of ` + count + `|a minimum of ` + count + `|no (?:more|longer) than ` + count + `) characters(?: in length| long)?`
)

const tagWord = `(?:Required|Optional|Computed|Sensitive|Deprecated|ForceNew|Read-Only|Write-Only|Boolean|Bool|String|Number|Integer|Int|Float|List|Set|Map|Block|Object)`

var (
	tag          = regexp.MustCompile(`\(` + tagWord + `(?:(?:, | and | or |/| )` + tagWord + `)*\)`)
	definedBelow = regexp.MustCompile(`[ \t]*\b(?:as )?(?:defined|documented|described|detailed|specified|outlined) (?:below|above)\b`)

	// whole sentences
	blockBelow     = regexp.MustCompile(`^(?:The|Each|Every|An?) (?:\x01+ )?blocks? (?:supports?|exports?|contains?|has|have)(?: the)?(?: following)?(?: fields| arguments| attributes| properties)?(?: as)? (?:defined|documented|described|detailed|specified|outlined|listed) (?:below|above)\.?$`)
	possibleValues = regexp.MustCompile(`^(?:(?:The )?(?i:possible|valid|allowed|accepted|supported|available) (?:values?|options) (?:are|is|include|includes|range)\b|(?:Possible|Valid|Allowed|Accepted|Supported|Available) (?:values|options):|Options (?:are|include)\b|(?:Must|Should|Can) be (?:one of|either|between|at least|at most)\b|(?:Can be (?:set to )?|One of |Either )` + values + `\.?$)`)
	defaults       = regexp.MustCompile(`^(?:(?:(?:If|When) (?:blank|omitted|unspecified|unset|not specified|not set|not provided|left blank|left empty),? )?(?:(?:This|this|It|it)(?: value| field| property)? |The |the )?[Dd]efaults? (?:value )?(?:to|is)\b|By default,? (?:this|it) is (?:set to )?` + value + `\.?$)`)
	forceNew       = regexp.MustCompile(`^Changing (?:this|the|it)\b(?: [\w\x01]+){0,3}? (?:forces|will force) a new\b`)
	oneOf          = regexp.MustCompile(`^(?:(?:Exactly |At least |Only )?[Oo]ne of (?:either )?|Either )` + values + ` (?:is required|(?:must|should|has to|have to|needs to|need to|can) be (?:specified|set|provided|configured|defined|used))\.?$`)
	conflictsWith  = regexp.MustCompile(`^(?:Conflicts with\b|Cannot be (?:set|used|specified|configured|defined) (?:together |along |alongside )?with\b|Only one of (?:either )?` + values + ` can be (?:specified|set|provided|configured|defined|used)\.?$)`)
	itemCount      = regexp.MustCompile(`^(?:(?:A |The )?(?i:maximum|minimum) of|(?:Can (?:support|have) |Supports )?(?i:up to|at most|at least|no more than)|Only) ` + count + ` (?:\x01+ )?(?:blocks?|items?|entries|elements?|values?)(?: (?:may|can|must|should) be (?:defined|specified|added|set|provided|configured)| (?:is|are) (?:allowed|supported))?\.?$`)
	lengthIs       = regexp.MustCompile(`^(?:(?:(?:It|This|The value|The name) )?(?i:must|should|can|may) be (?:a string )?` + span + `|Cannot exceed ` + count + ` characters(?: in length)?|Max(?:imum)? length(?: is|:) ` + count + ` characters|The maximum (?:allowed )?length (?:for this field )?is ` + count + ` characters|Length is limited to ` + count + ` characters|Restricted to ` + count + ` characters)\.?$`)

	// words inside a sentence that says more
	defaultClause  = regexp.MustCompile(`[,;] (?:which |and |it |this |otherwise )?defaults? to ` + value + `(\.?)$`)
	defaultBracket = regexp.MustCompile(` ?\((?:which |this |it )?[Dd]efaults? to ` + value + `\)`)
	lengthClause   = regexp.MustCompile(span)
)

// the validation that pins a property to a list or a range of values, by the name of the function that builds it
var (
	lengths    = []string{"StringLenBetween", "LengthBetween", "LengthAtLeast", "LengthAtMost", "UTF8LengthBetween", "UTF8LengthAtLeast", "UTF8LengthAtMost"}
	valueLists = []string{"StringInSlice", "IntInSlice", "FloatInSlice", "IntBetween", "IntAtLeast", "IntAtMost", "FloatBetween", "FloatAtLeast", "FloatAtMost", "OneOf", "OneOfCaseInsensitive", "Between", "AtLeast", "AtMost"}
)

// facts is what a property's schema says about it that a description might repeat.
type facts struct {
	values, isBool, hasDefault, forceNew, conflicts, oneOf, items, length bool
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	for _, found := range description.All(pass, insp) {
		var known facts
		if found.Owner != nil {
			fields := map[string]ast.Expr{}
			for _, elt := range found.Owner.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if key, ok := kv.Key.(*ast.Ident); ok {
						fields[key.Name] = kv.Value
					}
				}
			}

			// a list of strings is validated on its element
			validation := []ast.Expr{fields["ValidateFunc"], fields["ValidateDiagFunc"], fields["Validators"]}
			if elem, ok := ast.Unparen(fields["Elem"]).(*ast.UnaryExpr); ok {
				if lit, ok := elem.X.(*ast.CompositeLit); ok {
					for _, elt := range lit.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if key, ok := kv.Key.(*ast.Ident); ok && (key.Name == "ValidateFunc" || key.Name == "ValidateDiagFunc") {
								validation = append(validation, kv.Value)
							}
						}
					}
				}
			}

			calls := func(names []string, exprs ...ast.Expr) bool {
				called := false
				for _, expr := range exprs {
					if expr == nil {
						continue
					}

					ast.Inspect(expr, func(n ast.Node) bool {
						if call, ok := n.(*ast.CallExpr); ok {
							if fn := astx.CalledFunc(pass, call); fn != nil && slices.Contains(names, fn.Name()) {
								called = true
							}
						}

						return !called
					})
				}

				return called
			}

			known.values = calls(valueLists, validation...)
			known.length = calls(lengths, validation...)
			known.conflicts = fields["ConflictsWith"] != nil || calls([]string{"ConflictsWith"}, fields["Validators"])
			known.oneOf = fields["ExactlyOneOf"] != nil || fields["AtLeastOneOf"] != nil || calls([]string{"ExactlyOneOf", "AtLeastOneOf"}, fields["Validators"])
			known.items = fields["MaxItems"] != nil || fields["MinItems"] != nil || calls([]string{"SizeAtMost", "SizeAtLeast", "SizeBetween"}, fields["Validators"])
			known.forceNew = astx.IsTrueConstant(pass, fields["ForceNew"]) || calls([]string{"RequiresReplace"}, fields["PlanModifiers"])
			known.hasDefault = fields["Default"] != nil && !astx.IsNilValue(pass, fields["Default"])

			if sel, ok := ast.Unparen(fields["Type"]).(*ast.SelectorExpr); ok && sel.Sel.Name == "TypeBool" {
				known.isBool = true
			}

			t := pass.TypesInfo.TypeOf(found.Owner)
			if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
				t = ptr.Elem()
			}
			if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Name() == "BoolAttribute" {
				known.isBool = true
			}
		}

		for _, finding := range check(found.Text, known) {
			found.Report(pass, finding)
		}
	}

	return nil, nil
}

// check returns what a description repeats of the schema, in the order it appears.
func check(text string, known facts) []description.Finding {
	masked := description.Mask(text)

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

	// what a whole sentence can repeat, and whether the schema already says it
	repeated := []struct {
		kind, message string
		says          *regexp.Regexp
		known         bool
	}{
		{below, "description points at the docs page - a description has no page around it", blockBelow, true},
		{"one-of", "description says one of several must be set - the schema's ExactlyOneOf or AtLeastOneOf already does", oneOf, known.oneOf},
		{length, givesLength, lengthIs, known.length},
		{possible, "description lists the values it can take - the schema's validation already does", possibleValues, known.values},
		{possible, "description lists the values it can take - a bool is always true or false", possibleValues, known.isBool},
		{"default", givesDefault, defaults, known.hasDefault},
		{"force-new", "description says a change forces a new resource - the schema's ForceNew already does", forceNew, known.forceNew},
		{"conflicts-with", "description says what it conflicts with - the schema's ConflictsWith already does", conflictsWith, known.conflicts},
		{"item-count", "description says how many there can be - the schema's MaxItems or MinItems already does", itemCount, known.items},
	}

	sentences := description.Sentences(masked)
	repeats := make([][2]string, len(sentences))
	kept := 0
	for i, sentence := range sentences {
		for _, repeat := range repeated {
			if repeat.known && *enabled[repeat.kind] && repeat.says.MatchString(sentence.Text) {
				repeats[i] = [2]string{repeat.kind, repeat.message}
				break
			}
		}

		if repeats[i][0] == "" {
			kept++
		}
	}

	// a sentence goes with the space before it, or the space after it when nothing is kept in front; with no sentence left there is nothing
	// to fix it to
	before := false
	for i, sentence := range sentences {
		if repeats[i][0] != "" {
			finding := description.Finding{Start: sentence.Start, End: sentence.End, Message: repeats[i][1]}
			switch {
			case kept == 0:
			case before:
				finding.Start, finding.Fix = sentences[i-1].End, removeSentence
			case i+1 < len(sentences):
				finding.End, finding.Fix = sentences[i+1].Start, removeSentence
			}
			add(repeats[i][0], finding)

			continue
		}

		before = true

		// a sentence that is kept can still end with the default, which goes and leaves the full stop
		if at := defaultClause.FindStringSubmatchIndex(sentence.Text); at != nil && known.hasDefault {
			add("default", description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[2], Message: givesDefault, Fix: removeWords})
		}

		// a length in the middle of a sentence has no clean way out, so it is only pointed at
		if at := lengthClause.FindStringIndex(sentence.Text); at != nil && known.length {
			add(length, description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[1], Message: givesLength})
		}
	}

	if known.hasDefault {
		for _, at := range defaultBracket.FindAllStringIndex(masked, -1) {
			add("default", description.Finding{Start: at[0], End: at[1], Message: givesDefault, Fix: removeWords})
		}
	}

	for _, at := range tag.FindAllStringIndex(masked, -1) {
		start, end := at[0], at[1]
		finding := description.Finding{Start: start, End: end, Message: fmt.Sprintf("description has the tag %q - the schema already says so", masked[start:end]), Fix: "Remove the tag"}

		// the tag goes with one of the spaces beside it
		if trimmed := strings.TrimRight(masked[:start], " "); trimmed != "" && len(trimmed) < start {
			finding.Start = len(trimmed)
		} else {
			finding.End = len(masked) - len(strings.TrimLeft(masked[end:], " "))
		}
		add("tag", finding)
	}

	for _, at := range definedBelow.FindAllStringIndex(masked, -1) {
		add(below, description.Finding{Start: at[0], End: at[1], Message: fmt.Sprintf("description points at the docs page with %q - a description has no page around it", strings.TrimSpace(masked[at[0]:at[1]])), Fix: removeWords})
	}

	slices.SortStableFunc(findings, func(a, b description.Finding) int { return cmp.Compare(a.Start, b.Start) })

	return findings
}
