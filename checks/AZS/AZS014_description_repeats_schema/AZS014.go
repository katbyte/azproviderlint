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
)

// Analyzer checks descriptions for what the schema already states: the type, the default, the allowed values, that a change forces a new
// resource. The docs take those from the schema, so a description that repeats them says it twice and goes stale when the schema changes. On a
// read-only property there is nothing to set, so those are reported whatever the schema says.
var Analyzer = &analysis.Analyzer{
	Name:      "AZS014",
	Doc:       "check that resource and property descriptions do not repeat what the schema says",
	URL:       "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS014_description_repeats_schema/README.md",
	Requires:  []*analysis.Analyzer{description.Analyzer},
	FactTypes: []analysis.Fact{(*viaHelper)(nil)},
	Run:       run,
}

// enabled holds, by name, whether each kind of repeat is looked for. All are, until one is switched off.
var enabled = map[string]*bool{}

func init() {
	for _, kind := range [][2]string{
		{"tag", "a tag copied from the docs, such as (Optional) or (Boolean)"},
		{possible, "a sentence listing the values a property takes, when its validation or its type already does"},
		{"default", "a sentence or words giving the default, which is the schema's to say"},
		{"force-new", "a sentence saying a change forces a new resource, when the property is ForceNew"},
		{"conflicts-with", "a sentence saying what a property conflicts with, when it sets ConflictsWith"},
		{"one-of", "a sentence saying one of several properties must be set, when it sets ExactlyOneOf or AtLeastOneOf"},
		{"required-with", "a sentence saying what else must be set, when it sets RequiredWith"},
		{"item-count", "a sentence saying how many items or blocks there can be, when it sets MaxItems or MinItems"},
		{length, "a sentence or words giving how long a value can be, when its validation already does"},
		{below, "a sentence or words that point at the docs page, such as \"as defined below\""},
		{unchecked, "a sentence giving a range or a length, when nothing in the schema checks it"},
		{setting, "a read-only property described as something to set, such as \"Whether to...\""},
	} {
		enabled[kind[0]] = Analyzer.Flags.Bool(kind[0], true, "report "+kind[1]+" (false leaves it alone)")
	}
}

const (
	possible = "possible-values"
	length   = "length"
	below    = "defined-below"

	unchecked = "unchecked-range"
	setting   = "read-only-setting"

	listsValues  = "description lists the values it can take - the schema's validation already does"
	givesDefault = "description gives the default - the schema's Default already does"
	unsetDefault = "description gives the default - set Default in the schema or take it out; if it has to stay, say why in an //azignore"
	givesLength  = "description says how long it can be - the schema's validation already does"
	neverSet     = " - a read-only property cannot be set"

	removeSentence = "Remove the sentence"
	removeWords    = "Remove the words"
)

// what a value, a list of them and a count look like once the code in backticks is blanked out
const (
	value  = `(?:\x01+|'[^']*'|"[^"]*"|[\w-]+)`
	values = value + `(?:(?:, and |, or |, | and | or )` + value + `)+`
	quoted = `(?:\x01+|'[^']*'|"[^"]*")`
	names  = `\x01+(?:(?:, and |, or |, | and | or )\x01+)*`
	listed = `(?i:possibles?|valid|allowed|accepted|acceptable|supported|permitted|available) (?:values?|options)`
	noun   = `(?: property| properties| fields?| arguments?| attributes?| blocks?| values?)`
	given  = `(?:set|specified|provided|configured|defined|used)`
	count  = `(?:\x01+|\d+|one|two|three|four|five|six|seven|eight|nine|ten)`
	ranged = `(?:The )?(?:(?i:valid|allowed|supported|possible|accepted|permitted) )?(?i:range) is\b`
	span   = `(?:between ` + count + ` and ` + count + `|\d+-\d+|up to ` + count + `|at most ` + count + `|at least ` + count + `|a maximum of ` + count + `|a minimum of ` + count + `|no (?:more|longer) than ` + count + `) characters(?: in length| long)?`
)

const tagWord = `(?:Required|Optional|Computed|Sensitive|Deprecated|ForceNew|Read-Only|Write-Only|Boolean|Bool|String|Number|Integer|Int|Float|List|Set|Map|Block|Object)`

var (
	tag          = regexp.MustCompile(`\(` + tagWord + `(?:(?:, | and | or |/| )` + tagWord + `)*\)`)
	definedBelow = regexp.MustCompile(`[ \t]*\b(?:as )?(?:defined|documented|described|detailed|specified|outlined) (?:below|above)\b`)

	// whole sentences
	blockBelow     = regexp.MustCompile(`^(?:The|Each|Every|An?) (?:\x01+ )?blocks? (?:supports?|exports?|contains?|has|have)(?: the)?(?: following)?(?: fields| arguments| attributes| properties)?(?: as)? (?:defined|documented|described|detailed|specified|outlined|listed) (?:below|above)\.?$`)
	possibleValues = regexp.MustCompile(`^(?:(?:(?:At this time|Currently),? )?(?:[Tt]he )?(?:only )?` + listed + `(?::| (?:are|is|include|includes|range)\b)|Options (?:are|include)\b|(?:The )?Values (?:are|include):? ` + quoted + `|(?:(?:This|It|The)(?: value| number| field| property)? )?(?i:must|should|can) be (?:one of|either|between|at least|at most)\b|` + ranged + `|(?:[Cc]an be (?:set to )?|One of |Either )` + values + `\.?$|(?:Currently,? )?(?i:only) ` + quoted + ` (?:is|are) (?:currently )?(?:supported|allowed|valid|accepted)(?: currently| at this time)?\.?$)`)
	defaults       = regexp.MustCompile(`^(?:(?:(?:If|When) (?:blank|omitted|unspecified|unset|not specified|not set|not provided|left blank|left empty),? )?(?:(?:This|this|It|it)(?: value| field| property)? |The |the )?[Dd]efaults? (?:value )?(?:to|is)\b|By default,? (?:this|it) is (?:set to )?` + value + `\.?$)`)
	forceNew       = regexp.MustCompile(`^Changing (?:this|the|it)\b(?: [\w\x01]+){0,3}? (?:forces|will force) a new\b`)
	requiredWith   = regexp.MustCompile(`^(?:(?:(?:This|It)(?: value| field| property| argument| attribute| block| setting)? is )?(?i:required) (?:with|when|if) ` + names + `(?: (?:is|are) (?:also )?(?:set|specified|provided|configured|defined|used|enabled))?|(?:The )?` + names + noun + `? must (?:both |all )?be ` + given + ` together(?: or not at all)?|(?:If|When) (?:the )?` + names + noun + `? (?:is|are) ` + given + `,? (?:then )?(?:the )?` + names + noun + `? (?:(?:is|are) (?:also )?required|must (?:also )?be ` + given + `)|Requires ` + names + `(?: to be (?:set|specified|provided|configured|defined))?|` + names + ` must also be (?:set|specified|provided|configured|defined)|(?:Only valid|Can only be (?:set|used|specified)) (?:in conjunction |together |along )?with ` + names + `)\.?$`)
	oneOf          = regexp.MustCompile(`^(?:(?:(?:Exactly |At least |Only )?[Oo]ne of (?:either )?|Either )` + values + `|Either \x01+ or both ` + names + `) (?:is required|(?:must|should|has to|have to|needs to|need to|can) be (?:specified|set|provided|configured|defined|used))\.?$`)
	conflictsWith  = regexp.MustCompile(`^(?:Conflicts with\b|Cannot be (?:set|used|specified|configured|defined) (?:together |along |alongside )?with\b|Only one of (?:either )?` + values + ` can be (?:specified|set|provided|configured|defined|used)\.?$)`)
	itemCount      = regexp.MustCompile(`^(?:(?:A |The )?(?i:maximum|minimum) of|(?:Can (?:support|have) |Supports )?(?i:up to|at most|at least|no more than)|Only) ` + count + ` (?:\x01+ )?(?:blocks?|items?|entries|elements?|values?)(?: (?:may|can|must|should) be (?:defined|specified|added|set|provided|configured)| (?:is|are) (?:allowed|supported))?\.?$`)
	lengthIs       = regexp.MustCompile(`^(?:(?:(?:It|This|The value|The name) )?(?i:must|should|can|may) be (?:a string )?` + span + `|Cannot exceed ` + count + ` characters(?: in length)?|Max(?:imum)? length(?: is|:) ` + count + ` characters|The maximum (?:allowed )?length (?:for this field )?is ` + count + ` characters|Length is limited to ` + count + ` characters|Restricted to ` + count + ` characters)\.?$`)

	// words inside a sentence that says more
	valuesUnended  = regexp.MustCompile(`[^\s.?!,;:]( (?:Possibles?|Valid|Allowed|Accepted|Acceptable|Supported|Permitted|Available) (?:values?|options)(?::| (?:are|is|include|includes)\b).*?)\.?$`)
	valuesClause   = regexp.MustCompile(`(?:,| -| –|;) (?:it )?(?:(?:can|must|should) be (?:set to )?(?:either |one of |one or more (?:of|from) )?|either |one of |one or more (?:of|from) )(?:the following:? )?` + values + `(\.?)$`)
	listedClause   = regexp.MustCompile(`(?:,| -| –|;) (?:the )?` + listed + `(?:(?::| (?:are|is|include|includes|will be):?) (?:either |one of )?(?:` + values + `|` + quoted + `(?: (?:through|to) ` + quoted + `)?)| (?:can )?range (?:from|between) .*?)(\.?)$`)
	conditional    = regexp.MustCompile(`^(?i:if|when|otherwise|unless|for|while|where|with|in)\b`)
	defaultClause  = regexp.MustCompile(`[,;] (?:which |and |it |this |otherwise )?defaults? to ` + value + `(\.?)$`)
	defaultBracket = regexp.MustCompile(` ?\((?:which |this |it )?[Dd]efaults? to ` + value + `\)`)
	lengthClause   = regexp.MustCompile(span)

	// what a description of something to set opens with: "Whether to...", "Set to...", "Whether the Thing should..."
	asSetting = regexp.MustCompile(`^\s*(?:\(\w+\) )?(?:(Whether to|Set|Enable|Disable) |Whether [^.?!]*?\b(should) be\b)`)

	// a sentence giving the range a value must be in
	rangeIs = regexp.MustCompile(`^(?:(?:(?:This|It|The)(?: value| number| field| property)? )?(?i:must|should|can) be (?:between|at least|at most)\b|(?:The )?` + listed + ` (?:(?:are|is) (?:between|from)|range (?:between|from))\b|` + ranged + `)`)
)

// the validation that pins a property to a length, and to a list or a range of values, by the name of the function that builds it
var (
	lengths    = []string{"StringLenBetween", "LengthBetween", "LengthAtLeast", "LengthAtMost", "UTF8LengthBetween", "UTF8LengthAtLeast", "UTF8LengthAtMost"}
	valueLists = []string{"StringInSlice", "IntInSlice", "FloatInSlice", "IntBetween", "IntAtLeast", "IntAtMost", "FloatBetween", "FloatAtLeast", "FloatAtMost", "OneOf", "OneOfCaseInsensitive", "Between", "AtLeast", "AtMost", "ISO8601DurationBetween"}
)

// formats holds the validation that checks what a value looks like and nothing of its range, with the validation to use when it has one.
var formats = map[string]string{"ISO8601Duration": "ISO8601DurationBetween", "StringIsNotEmpty": "", "StringIsNotWhiteSpace": "", "NoZeroValues": ""}

// facts is what a property's schema says about it that a description might repeat.
type facts struct {
	Values, IsBool, HasDefault, ForceNew, Conflicts, OneOf, Requires, Items, Length bool

	// ReadOnly is whether the provider fills the property in and the user cannot set it: Computed, and neither Optional nor Required.
	ReadOnly bool

	// Unchecked is whether nothing validates the property, or only its format, so a range in its description is not enforced; Bounds names the
	// validation that would enforce it, when the schema tells which
	Unchecked bool
	Bounds    string
}

// viaHelper is the fact on a function that builds a schema around a description it is given: what that schema says, by the parameter the
// description comes in as. It is how a description handed over from another package is checked against the schema it ends up on.
type viaHelper struct{ Params map[int]facts }

func (*viaHelper) AFact() {}

func (v *viaHelper) String() string { return fmt.Sprintf("schema(%d)", len(v.Params)) }

func run(pass *analysis.Pass) (any, error) {
	descriptions, ok := pass.ResultOf[description.Analyzer].(*description.Descriptions)
	if !ok {
		return nil, nil
	}

	// what a schema value says about its property; nothing when there is none to read
	read := func(owner *ast.CompositeLit) facts {
		var known facts
		if owner == nil {
			return known
		}

		fields := map[string]ast.Expr{}
		for _, elt := range owner.Elts {
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

		known.Values = calls(valueLists, validation...)
		known.Length = calls(lengths, validation...)
		// exactly one of several also means no two of them together
		known.Conflicts = fields["ConflictsWith"] != nil || fields["ExactlyOneOf"] != nil || calls([]string{"ConflictsWith", "ExactlyOneOf"}, fields["Validators"])
		known.Requires = fields["RequiredWith"] != nil || calls([]string{"AlsoRequires"}, fields["Validators"])
		known.OneOf = fields["ExactlyOneOf"] != nil || fields["AtLeastOneOf"] != nil || calls([]string{"ExactlyOneOf", "AtLeastOneOf"}, fields["Validators"])
		known.Items = fields["MaxItems"] != nil || fields["MinItems"] != nil || calls([]string{"SizeAtMost", "SizeAtLeast", "SizeBetween"}, fields["Validators"])
		known.ForceNew = astx.IsTrueConstant(pass, fields["ForceNew"]) || calls([]string{"RequiresReplace"}, fields["PlanModifiers"])
		known.HasDefault = fields["Default"] != nil && !astx.IsNilValue(pass, fields["Default"])
		known.ReadOnly = astx.IsTrueConstant(pass, fields["Computed"]) && fields["Optional"] == nil && fields["Required"] == nil

		if sel, ok := ast.Unparen(fields["Type"]).(*ast.SelectorExpr); ok {
			known.IsBool = sel.Sel.Name == "TypeBool"
			known.Bounds = map[string]string{"TypeInt": "IntBetween", "TypeFloat": "FloatBetween"}[sel.Sel.Name]
		}

		// a function this check does not know may well check the range, so only no validation, or one that is known not to, counts
		known.Unchecked = true
		for _, expr := range validation {
			name := ""
			switch v := ast.Unparen(expr).(type) {
			case nil:
				continue
			case *ast.Ident:
				name = v.Name
			case *ast.SelectorExpr:
				name = v.Sel.Name
			}

			better, ok := formats[name]
			known.Unchecked = known.Unchecked && ok
			if better != "" {
				known.Bounds = better
			}
		}

		t := pass.TypesInfo.TypeOf(owner)
		if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
			t = ptr.Elem()
		}
		if named, ok := types.Unalias(t).(*types.Named); ok && named.Obj().Name() == "BoolAttribute" {
			known.IsBool = true
		}

		return known
	}

	// what each of this package's helpers says about the property it builds, for the packages that hand it a description
	for fn, params := range descriptions.Helpers {
		via := viaHelper{Params: map[int]facts{}}
		for i, owner := range params {
			via.Params[i] = read(owner)
		}
		pass.ExportObjectFact(fn, &via)
	}

	for _, found := range descriptions.Found {
		known := read(found.Owner)

		var via viaHelper
		if found.Owner == nil && found.Via != nil && pass.ImportObjectFact(found.Via, &via) {
			known = via.Params[found.Param]
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

	// nothing is set on a read-only property, so what a description says can be set on it is reported and taken out as if the schema said it too
	if known.ReadOnly {
		known = facts{ReadOnly: true, Values: true, HasDefault: true, ForceNew: true, Conflicts: true, OneOf: true, Requires: true, Items: true, Length: true}
	}

	// a later finding that overlaps an earlier one is dropped, so no two fixes rewrite the same bytes
	var findings []description.Finding
	add := func(kind string, finding description.Finding) {
		if !*enabled[kind] {
			return
		}

		if what, _, ok := strings.Cut(finding.Message, " - "); ok && known.ReadOnly && kind != "tag" && kind != below && kind != setting {
			finding.Message = what + neverSet
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
		{"one-of", "description says one of several must be set - the schema's ExactlyOneOf or AtLeastOneOf already does", oneOf, known.OneOf},
		{length, givesLength, lengthIs, known.Length},
		{possible, listsValues, possibleValues, known.Values},
		{possible, "description lists the values it can take - a bool is always true or false", possibleValues, known.IsBool},
		{"default", givesDefault, defaults, known.HasDefault},
		{"force-new", "description says a change forces a new resource - the schema's ForceNew already does", forceNew, known.ForceNew},
		{"conflicts-with", "description says what it conflicts with - the schema's ConflictsWith or ExactlyOneOf already does", conflictsWith, known.Conflicts},
		{"required-with", "description says what else must be set - the schema's RequiredWith already does", requiredWith, known.Requires},
		{"item-count", "description says how many there can be - the schema's MaxItems or MinItems already does", itemCount, known.Items},
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

		// a range belongs in the schema, where it is enforced; the sentence stays until it is
		if known.Unchecked && (rangeIs.MatchString(sentence.Text) || lengthIs.MatchString(sentence.Text)) {
			with := "a validation"
			switch {
			case lengthIs.MatchString(sentence.Text):
				with = "StringLenBetween"
			case known.Bounds != "":
				with = known.Bounds
			}
			add(unchecked, description.Finding{Start: sentence.Start, End: sentence.End, Message: "description gives a range the schema does not check - check it in the schema with " + with})
		}

		// a sentence that is kept can still end with the values, after a comma or where a full stop was left out
		if known.Values {
			if at := valuesClause.FindStringSubmatchIndex(sentence.Text); at != nil {
				add(possible, description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[2], Message: listsValues, Fix: removeWords})
			}

			if at := valuesUnended.FindStringSubmatchIndex(sentence.Text); at != nil {
				add(possible, description.Finding{Start: sentence.Start + at[2], End: sentence.Start + at[3], Message: listsValues, Fix: removeWords})
			}

			// "If mode is Fast, possible values are..." says when, which the schema does not
			if at := listedClause.FindStringSubmatchIndex(sentence.Text); at != nil && !conditional.MatchString(sentence.Text) {
				add(possible, description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[2], Message: listsValues, Fix: removeWords})
			}
		}

		// a default is called out wherever it is written, but only taken out when the schema gives it too: with no Default there, taking it out
		// would lose the only place it is written down
		switch at := defaultClause.FindStringSubmatchIndex(sentence.Text); {
		case !known.HasDefault && defaults.MatchString(sentence.Text):
			add("default", description.Finding{Start: sentence.Start, End: sentence.End, Message: unsetDefault})

		case at != nil && known.HasDefault:
			add("default", description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[2], Message: givesDefault, Fix: removeWords})

		case at != nil:
			add("default", description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[2], Message: unsetDefault})
		}

		// a length in the middle of a sentence has no clean way out, so it is only pointed at
		if at := lengthClause.FindStringIndex(sentence.Text); at != nil && known.Length {
			add(length, description.Finding{Start: sentence.Start + at[0], End: sentence.Start + at[1], Message: givesLength})
		}
	}

	for _, at := range defaultBracket.FindAllStringIndex(masked, -1) {
		if !known.HasDefault {
			add("default", description.Finding{Start: at[0], End: at[1], Message: unsetDefault})

			continue
		}

		// at the very start there is no space in front to take, so the one behind goes
		finding := description.Finding{Start: at[0], End: at[1], Message: givesDefault, Fix: removeWords}
		if at[0] == 0 {
			finding.End = len(masked) - len(strings.TrimLeft(masked[at[1]:], " "))
		}
		add("default", finding)
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

	// a read-only property is filled in, not set, so its description says what it holds
	if at := asSetting.FindStringSubmatchIndex(masked); at != nil && known.ReadOnly {
		start, end := max(at[2], at[4]), max(at[3], at[5])
		add(setting, description.Finding{Start: start, End: end, Message: fmt.Sprintf("description reads as something to set with %q - a read-only property cannot be set, say what it holds", masked[start:end])})
	}

	for _, at := range definedBelow.FindAllStringIndex(masked, -1) {
		add(below, description.Finding{Start: at[0], End: at[1], Message: fmt.Sprintf("description points at the docs page with %q - a description has no page around it", strings.TrimSpace(masked[at[0]:at[1]])), Fix: removeWords})
	}

	slices.SortStableFunc(findings, func(a, b description.Finding) int { return cmp.Compare(a.Start, b.Start) })

	return findings
}
