// Package AZG014 defines an analyzer that reports an error message that is not formatted the way the provider's others are.
package AZG014

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks how the errors a provider makes are formatted, so a user reads every one the same way: starting in lower case, the error
// underneath printed in full and words written out. It can also ask for values in backticks, as the docs have them.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG014",
	Doc:      "check that error messages are formatted the way the provider's others are",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG014_error_message_formatting/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// each kind of wording can be switched off on its own
var capital, errorVerb, contraction, quotedValue bool

// names are the words a message may start with a capital for. Words given as options are added to these.
var names = []string{"Azure", "Microsoft", "Terraform"}

func init() {
	Analyzer.Flags.Init("AZG014", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&capital, "capital", true, "report a message that starts with a capital letter (false leaves it alone)")
	Analyzer.Flags.Func("name", "a name a message may start with a capital for, such as Azure; repeat for each, which adds to the ones the rule knows", func(name string) error {
		names = append(names, name)

		return nil
	})
	Analyzer.Flags.BoolVar(&errorVerb, "error-verb", true, "report an error printed with %s or %v, which should be %+v (false leaves it alone)")
	Analyzer.Flags.BoolVar(&contraction, "contraction", true, "report a contraction such as can't (false leaves it alone)")
	Analyzer.Flags.BoolVar(&quotedValue, "quoted-value", false, "also report a value put in quotes with %q, which the docs put in backticks")
}

// written is each contraction this rule knows, and how it is written out.
var written = map[string]string{"can't": "cannot", "won't": "will not", "don't": "do not", "doesn't": "does not", "didn't": "did not", "isn't": "is not", "aren't": "are not", "wasn't": "was not", "weren't": "were not", "couldn't": "could not", "shouldn't": "should not", "wouldn't": "would not", "hasn't": "has not", "haven't": "have not", "mustn't": "must not"}

// opening is the word a message starts with when that is a capital and then only small letters, which leaves alone a short form such as ID or
// IPv4 and a name from the code such as DecodeDiff.
var opening = regexp.MustCompile(`^([A-Z][a-z']+)(?:\W|$)`)

var contracted = regexp.MustCompile(`(?i)\b(?:[a-z]+n't|it's)\b`)

// verb is one place in a message where a value is printed: the letter, anything written between the % and it, and which value it prints.
var verb = regexp.MustCompile(`%([-+# 0]*)([0-9*]*)(\.[0-9*]*)?(\[\d+\])?([a-zA-Z%])`)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	failure, isFailure := types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
	if !isFailure {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, isCall := n.(*ast.CallExpr)
		if !isCall || len(call.Args) == 0 || strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return
		}

		// the two ways an error is made from a message
		fn := astx.CalledFunc(pass, call)
		if fn == nil || fn.Pkg() == nil || (fn.FullName() != "fmt.Errorf" && fn.FullName() != "errors.New") {
			return
		}

		message, isText := pass.TypesInfo.Types[call.Args[0]]
		if !isText || message.Value == nil || message.Value.Kind() != constant.String {
			return
		}

		// a place in the message is a place in the code when the message is one string written with nothing escaped in it
		text := constant.StringVal(message.Value)
		lit, isLit := ast.Unparen(call.Args[0]).(*ast.BasicLit)
		plain := isLit && len(lit.Value) == len(text)+2 && lit.Value[1:len(lit.Value)-1] == text
		report := func(start, end int, message, fix, replacement string) {
			d := analysis.Diagnostic{Pos: call.Args[0].Pos(), End: call.Args[0].End(), Message: message}
			if plain {
				d.Pos, d.End = lit.Pos()+token.Pos(1+start), lit.Pos()+token.Pos(1+end)
			}
			if plain && fix != "" {
				d.SuggestedFixes = []analysis.SuggestedFix{{Message: fix, TextEdits: []analysis.TextEdit{{Pos: d.Pos, End: d.End, NewText: []byte(replacement)}}}}
			}
			pass.Report(d)
		}

		// an error is read after what it happened in, as in "creating Thing: expected a name"
		// a contraction that starts a message is written out in lower case below, in one fix
		if word := opening.FindStringSubmatch(text); capital && word != nil && !slices.Contains(names, word[1]) && written[strings.ToLower(word[1])] == "" {
			report(0, len(word[1]), "error message starts with a capital - write \""+strings.ToLower(word[1])+"\"", "start it in lower case", strings.ToLower(word[1]))
		}

		if contraction {
			for _, at := range contracted.FindAllStringIndex(text, -1) {
				// "it's" is "it is", "it has", or "its" spelt wrong, which is for a person to say
				word := text[at[0]:at[1]]
				if strings.EqualFold(word, "it's") {
					report(at[0], at[1], "error message has \""+word+"\" - write \"it is\", \"it has\" or \"its\"", "", "")
					continue
				}

				out, isKnown := written[strings.ToLower(word)]
				if !isKnown {
					continue
				}

				if word[0] != out[0] && !capital {
					out = strings.ToUpper(out[:1]) + out[1:]
				}
				report(at[0], at[1], "error message has \""+word+"\" - write \""+out+"\"", "write it out", out)
			}
		}

		if fn.Name() != "Errorf" {
			return
		}

		// the values are printed in the order given, unless a place says which one it prints, and then this rule cannot follow
		next := 1
		for _, at := range verb.FindAllStringSubmatchIndex(text, -1) {
			letter := text[at[10]:at[11]]
			if at[8] >= 0 {
				return
			}
			if letter == "%" {
				continue
			}

			next += strings.Count(text[at[4]:at[5]], "*")
			if at[6] >= 0 {
				next += strings.Count(text[at[6]:at[7]], "*")
			}

			if next >= len(call.Args) || call.Ellipsis.IsValid() {
				return
			}

			value := pass.TypesInfo.TypeOf(call.Args[next])
			next++
			if value == nil {
				continue
			}

			bare := at[11]-at[0] == 2
			switch {
			case errorVerb && bare && letter == "v" && strings.HasSuffix(text[:at[0]], " +"):
				report(at[0]-1, at[1], "error message has +%v, with the plus on the wrong side - use %+v", "print it with %+v", "%+v")

			case errorVerb && bare && (letter == "s" || letter == "v") && types.Implements(value, failure):
				report(at[0], at[1], "error is printed with %"+letter+" - use %+v, which keeps everything the error says", "print it with %+v", "%+v")

			case quotedValue && bare && letter == "q" && types.Identical(value.Underlying(), types.Typ[types.String]) && (!isLit || lit.Value[0] == '"'):
				report(at[0], at[1], "error message puts a value in quotes with %q - put it in backticks, as the docs do", "put it in backticks", "`%s`")
			}
		}
	})

	return nil, nil
}
