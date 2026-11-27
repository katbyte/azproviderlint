// Package AZS028 defines an analyzer that reports a list or set property whose name is not singular for a block or plural for values.
package AZS028

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks what a list or set property's name says about it: a block the user writes once per thing has a singular name, and a list of
// plain values has a plural one.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS028",
	Doc:      "check that a list or set of blocks has a singular name, and a list or set of values a plural one",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS028_list_property_names/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// each kind of name can be switched off on its own
var blockSingular, listPlural bool

// singular are words ending in s that are not plurals, and plural are plurals the ending does not show. A word in both, such as data, has no
// other form and is fine as either. Words given as options are added to these.
var singular, plural = []string{"alias", "analytics", "aws", "contains", "cors", "data", "diagnostics", "dns", "equals", "https", "insights", "kms", "kubernetes", "list", "media", "metadata", "nfs", "os", "redis", "sas", "series", "tls", "windows"}, []string{"apis", "cpus", "criteria", "data", "gpus", "list", "media", "metadata", "uris", "vcpus"}

// trailing are the words that may come after the plural in a list's name, as in days_of_week. Words given as options are added to these.
var trailing = []string{"after", "allowed", "base64", "for", "from", "in", "not", "of", "only", "to", "with"}

func init() {
	Analyzer.Flags.Init("AZS028", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&blockSingular, "block-singular", true, "report a block with a plural name (false leaves it alone)")
	Analyzer.Flags.BoolVar(&listPlural, "list-plural", true, "report a list of values with a singular name (false leaves it alone)")
	Analyzer.Flags.Func("singular-word", "a word ending in s that is not a plural, such as dns; repeat for each, which adds to the ones the rule knows", func(word string) error {
		singular = append(singular, word)

		return nil
	})
	Analyzer.Flags.Func("plural-word", "a plural that does not end in s, such as data; repeat for each, which adds to the ones the rule knows", func(word string) error {
		plural = append(plural, word)

		return nil
	})
	Analyzer.Flags.Func("trailing-word", "a word that may come after the plural in a list's name, such as of in days_of_week; repeat for each, which adds to the ones the rule knows", func(word string) error {
		trailing = append(trailing, word)

		return nil
	})
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, and its names cannot change
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	// a property is named by its key in the map it sits in, which is met before the property is
	keys := map[*ast.CompositeLit]ast.Expr{}
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || pass.TypesInfo.TypeOf(lit) == nil {
			return
		}

		fields, inside := map[string]ast.Expr{}, map[string]*ast.CompositeLit{}
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := kv.Value
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}

			inner, isInner := value.(*ast.CompositeLit)
			if isInner {
				keys[inner] = kv.Key
			}
			if key, isField := kv.Key.(*ast.Ident); isField {
				fields[key.Name], inside[key.Name] = kv.Value, inner
			}
		}

		t := pass.TypesInfo.TypeOf(lit)
		if ptr, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
			t = ptr.Elem()
		}

		named, isNamed := types.Unalias(t).(*types.Named)
		key, isName := pass.TypesInfo.Types[keys[lit]]
		if !isNamed || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" || !isName || key.Value == nil || key.Value.Kind() != constant.String {
			return
		}

		// a name on its way out is not worth changing
		if fields["Deprecated"] != nil || fields["DeprecationMessage"] != nil {
			return
		}

		name, kind := constant.StringVal(key.Value), types.ExprString(fields["Type"])
		words := strings.Split(name, "_")
		last := words[len(words)-1]

		// a list's name is plural when its last word is, or the word before one that may trail it, as in days_of_week
		many := false
		for i, word := range words {
			if i == len(words)-1 || slices.Contains(trailing, words[i+1]) {
				many = many || slices.Contains(plural, word) || (strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") && !strings.HasSuffix(word, "us") && !strings.HasSuffix(word, "is") && !slices.Contains(singular, word))
			}
		}
		collection := named.Obj().Name() == "Schema" && (strings.HasSuffix(kind, "TypeList") || strings.HasSuffix(kind, "TypeSet")) && inside["Elem"] != nil
		switch {
		// a block the user writes is one thing each time it is written; one that is only read back is a list of them
		case blockSingular && collection && (slices.Contains(plural, last) || (strings.HasSuffix(last, "s") && !strings.HasSuffix(last, "ss") && !strings.HasSuffix(last, "us") && !strings.HasSuffix(last, "is"))) && !slices.Contains(singular, last) && tf.IsSchemaHelperType(pass, inside["Elem"], "Resource") && (astx.IsTrueConstant(pass, fields["Required"]) || astx.IsTrueConstant(pass, fields["Optional"])):
			pass.Reportf(keys[lit].Pos(), "block %s has a plural name - a block is one thing each time it is written, so name it in the singular", strconv.Quote(name))

		case listPlural && collection && !many && tf.IsSchemaHelperType(pass, inside["Elem"], "Schema"):
			pass.Reportf(keys[lit].Pos(), "list %s has a singular name - it holds many values, so name it in the plural", strconv.Quote(name))
		}
	})

	return nil, nil
}
