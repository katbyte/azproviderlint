// Package AZG014 defines an analyzer that reports an error message that is not worded the way the provider's others are.
package AZG014

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks the wording of the errors a provider makes, so a user reads every one the same way: the error underneath printed in full,
// words written out, and values in backticks as the docs have them.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG014",
	Doc:      "check that error messages print the error underneath in full, have no contractions, and put values in backticks",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG014_error_message_wording/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// each kind of wording can be switched off on its own
var errorVerb, contraction, quotedValue *bool

func init() {
	Analyzer.Flags.Init("AZG014", flag.ContinueOnError)
	errorVerb = Analyzer.Flags.Bool("error-verb", true, "report an error printed with %s or %v, which should be %+v (false leaves it alone)")
	contraction = Analyzer.Flags.Bool("contraction", true, "report a contraction such as can't (false leaves it alone)")
	quotedValue = Analyzer.Flags.Bool("quoted-value", true, "report a value put in quotes with %q, which should be in backticks (false leaves it alone)")
}

// written is each contraction this rule knows, and how it is written out.
var written = map[string]string{"can't": "cannot", "won't": "will not", "don't": "do not", "doesn't": "does not", "didn't": "did not", "isn't": "is not", "aren't": "are not", "wasn't": "was not", "weren't": "were not", "couldn't": "could not", "shouldn't": "should not", "wouldn't": "would not", "hasn't": "has not", "haven't": "have not", "mustn't": "must not"}

var contracted = regexp.MustCompile(`(?i)\b[a-z]+n't\b`)

// verb is one place in a message where a value is printed: the letter, anything written between the % and it, and which value it prints.
var verb = regexp.MustCompile(`%([-+# 0]*)([0-9*]*)(\.[0-9*]*)?(\[[0-9]+\])?([a-zA-Z%])`)

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
				d.SuggestedFixes = []analysis.SuggestedFix{{Message: fix, TextEdits: []analysis.TextEdit{{Pos: d.Pos, End: d.End, NewText: []byte(replacement)}}}}
			}
			pass.Report(d)
		}

		if *contraction {
			for _, at := range contracted.FindAllStringIndex(text, -1) {
				word := text[at[0]:at[1]]
				out, isKnown := written[strings.ToLower(word)]
				if !isKnown {
					continue
				}

				if word[0] != out[0] {
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
			case *errorVerb && bare && (letter == "s" || letter == "v") && types.Implements(value, failure):
				report(at[0], at[1], "error is printed with %"+letter+" - use %+v, which keeps everything the error says", "print it with %+v", "%+v")

			case *quotedValue && bare && letter == "q" && types.Identical(value.Underlying(), types.Typ[types.String]) && (!isLit || lit.Value[0] == '"'):
				report(at[0], at[1], "error message puts a value in quotes with %q - put it in backticks, as the docs do", "put it in backticks", "`%s`")
			}
		}
	})

	return nil, nil
}
