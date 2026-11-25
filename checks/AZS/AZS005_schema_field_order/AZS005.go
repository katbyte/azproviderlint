// Package AZS005 defines an analyzer that reports schema properties whose fields are not written in the agreed order.
package AZS005

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that the fields of a schema property are written in one order everywhere, so every property reads the same way: what it is, its
// type, whether it is set, then how it is checked, with what it holds last.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS005",
	Doc:      "check that the fields of a schema property are written in the agreed order",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS005_schema_field_order/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// frameworkOrder is the order the fields of a framework attribute or block are written in, and pluginOrder that of a plugin SDK schema: the one
// most properties in the azurerm, azuread and aws providers already use, with what the property holds last. A field in neither stays where it is.
var (
	frameworkOrder = []string{"Description", "MarkdownDescription", "CustomType", "ElementType", "NestedObject", "AttributeTypes", "Required", "Optional", computed, "Default", "Sensitive", "WriteOnly", "Validators", "PlanModifiers", "DeprecationMessage", "Attributes", "Blocks"}
	pluginOrder    = []string{"Description", "Type", "Required", "Optional", computed, "ForceNew", "Default", "DefaultFunc", "InputDefault", "Sensitive", "WriteOnly", "MinItems", "MaxItems", "ConfigMode", "ValidateFunc", "ValidateDiagFunc", "StateFunc", "DiffSuppressFunc", "DiffSuppressOnRefresh", "Set", "RequiredWith", "ConflictsWith", "ExactlyOneOf", "AtLeastOneOf", "Deprecated", "Elem"}
)

const computed = "Computed"

// stateMigrations is whether the schemas in a state migration are checked like any other.
var stateMigrations = true

// bothNote is the comment AZS007 asks for on a property that is both Optional and Computed.
var bothNote = regexp.MustCompile(`(?i)//\s*note:\s*o\+c`)

func init() {
	Analyzer.Flags.Init("AZS005", flag.ContinueOnError)

	// the first field given for a list replaces its default, and the rest add to it
	given := map[*[]string]bool{}
	list := func(into *[]string) func(string) error {
		return func(field string) error {
			if !token.IsIdentifier(field) || (given[into] && slices.Contains(*into, field)) {
				return fmt.Errorf("%q is not a field name, or is given twice", field)
			}

			if !given[into] {
				*into, given[into] = nil, true
			}
			*into = append(*into, field)

			return nil
		}
	}

	// order and order-framework are two names for the one setting
	for _, name := range []string{"order", "order-framework"} {
		Analyzer.Flags.Func(name, "a field of a framework attribute or block, in the order the fields are written; repeat for each, which replaces the default order", list(&frameworkOrder))
	}

	Analyzer.Flags.Func("order-plugin", "a field of a plugin SDK schema, in the order the fields are written; repeat for each, which replaces the default order", list(&pluginOrder))
	Analyzer.Flags.BoolVar(&stateMigrations, "state-migrations", true, "also check state migrations, whose schemas are what old state looked like (false leaves them alone)")
}

// property is one schema property written out in the code, and the order its fields go in.
type property struct {
	lit  *ast.CompositeLit
	want []string
}

// found is one property with a field out of order. Fix is the whole property rewritten in order, and empty when it is laid out in a way that
// cannot be moved line by line, or when the rewrite of a property around it already holds it.
type found struct {
	lit     *ast.CompositeLit
	at      *ast.Ident
	message string
	fix     string
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok || (!stateMigrations && tf.InStateMigration(pass)) {
		return nil, nil
	}

	// the plugin sdk and the framework both keep their schema types in a package named schema
	var properties []property
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return
		}

		t := pass.TypesInfo.TypeOf(lit)
		if t == nil {
			return
		}
		if ptr, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
			t = ptr.Elem()
		}

		named, isNamed := types.Unalias(t).(*types.Named)
		if !isNamed || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" {
			return
		}

		switch name := named.Obj().Name(); {
		case name == "Schema":
			properties = append(properties, property{lit: lit, want: pluginOrder})
		case strings.HasSuffix(name, "Attribute") || strings.HasSuffix(name, "Block"):
			properties = append(properties, property{lit: lit, want: frameworkOrder})
		}
	})

	for _, f := range check(pass.Fset, pass.ReadFile, properties) {
		d := analysis.Diagnostic{Pos: f.at.Pos(), End: f.at.End(), Message: f.message}
		if f.fix != "" {
			d.SuggestedFixes = []analysis.SuggestedFix{{Message: "Put the fields in order", TextEdits: []analysis.TextEdit{{Pos: f.lit.Lbrace, End: f.lit.Rbrace + 1, NewText: []byte(f.fix)}}}}
		}
		pass.Report(d)
	}

	return nil, nil
}

// check returns the properties with a field out of order, in the order they are written. The properties come as they are met walking the code, so
// one around another comes before it.
func check(fset *token.FileSet, read func(string) ([]byte, error), properties []property) []found {
	var all []found

	// the innermost first, so the rewrite of a property can take in those of the properties inside it
	for _, p := range slices.Backward(properties) {
		lit, n := p.lit, len(p.lit.Elts)

		var keys []*ast.Ident
		for _, elt := range lit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if key, ok := kv.Key.(*ast.Ident); ok {
					keys = append(keys, key)
				}
			}
		}
		if len(keys) != n {
			continue
		}

		// the fields the order names are sorted among the places they hold now; any other field stays where it is
		var places []int
		for i, key := range keys {
			if slices.Contains(p.want, key.Name) {
				places = append(places, i)
			}
		}

		sorted := slices.Clone(places)
		slices.SortStableFunc(sorted, func(a, b int) int { return slices.Index(p.want, keys[a].Name) - slices.Index(p.want, keys[b].Name) })
		if slices.Equal(sorted, places) {
			continue
		}

		now := make([]int, n)
		for i := range now {
			now[i] = i
		}
		f := found{lit: lit}
		for i, place := range places {
			now[place] = sorted[i]
			if f.at == nil && sorted[i] != place {
				f.at, f.message = keys[sorted[i]], fmt.Sprintf("schema field %s should come before %s", keys[sorted[i]].Name, keys[place].Name)
			}
		}
		all = append(all, f)
		this := len(all) - 1

		tfile := fset.File(lit.Pos())
		src, err := read(tfile.Name())
		if err != nil {
			continue
		}

		// a field is moved as what it is when the property is on one line, and as its lines, with any comment above it, when each field has its own
		starts, ends, movable := make([]int, n), make([]int, n), true
		for i, elt := range lit.Elts {
			starts[i], ends[i] = tfile.Offset(elt.Pos()), tfile.Offset(elt.End())
			if tfile.Line(lit.Lbrace) == tfile.Line(lit.Rbrace) {
				continue
			}

			after := lit.Lbrace
			if i > 0 {
				after = lit.Elts[i-1].End()
			}
			if tfile.Line(elt.Pos()) == tfile.Line(after) || tfile.Line(elt.End()) == tfile.Line(lit.Rbrace) {
				movable = false
				break
			}
			starts[i], ends[i] = tfile.Offset(tfile.LineStart(tfile.Line(after)+1)), tfile.Offset(tfile.LineStart(tfile.Line(elt.End())+1))
		}
		if !movable {
			continue
		}

		// a property inside a field goes in already rewritten, since two fixes cannot rewrite the same lines. Those found so far run from the end
		// of the file back, so putting one in does not move the place of the next.
		texts, inside := make([]string, n), []int{}
		for i := range lit.Elts {
			texts[i] = string(src[starts[i]:ends[i]])
			for j, inner := range all {
				from, to := tfile.Offset(inner.lit.Lbrace), tfile.Offset(inner.lit.Rbrace)+1
				if inner.fix == "" || fset.File(inner.lit.Pos()) != tfile || from < starts[i] || to > ends[i] {
					continue
				}
				texts[i] = texts[i][:from-starts[i]] + inner.fix + texts[i][to-starts[i]:]
				inside = append(inside, j)
			}
		}

		// the note saying why a property is both set and read back sits between Optional and Computed, whichever field it was written above
		if to := slices.IndexFunc(keys, func(key *ast.Ident) bool { return key.Name == computed }); to >= 0 && tfile.Line(lit.Lbrace) != tfile.Line(lit.Rbrace) {
			for i, elt := range lit.Elts {
				above := tfile.Offset(tfile.LineStart(tfile.Line(elt.Pos()))) - starts[i]
				if i != to && bothNote.MatchString(texts[i][:above]) {
					texts[to], texts[i] = texts[i][:above]+texts[to], texts[i][above:]
				}
			}
		}

		// gofmt lines the values up afresh, which it only does for code it can read: the property is given a name to stand in for its type, and
		// the indent of its first line
		line := string(src[tfile.Offset(tfile.LineStart(tfile.Line(lit.Lbrace))):tfile.Offset(lit.Lbrace)])
		opening := line[:len(line)-len(strings.TrimLeft(line, " \t"))] + "_ = T"

		var text strings.Builder
		text.WriteString(opening)
		text.Write(src[tfile.Offset(lit.Lbrace):starts[0]])
		for i, field := range now {
			// a blank line above a field marked a break in the old order, and is left out of the new one
			for line, rest, more := strings.Cut(texts[field], "\n"); more && strings.TrimSpace(line) == ""; line, rest, more = strings.Cut(rest, "\n") {
				texts[field] = rest
			}

			text.WriteString(texts[field])
			if i < n-1 {
				text.Write(src[ends[i]:starts[i+1]])
			}
		}
		text.Write(src[ends[n-1] : tfile.Offset(lit.Rbrace)+1])

		formatted, err := format.Source([]byte(text.String()))
		if err != nil || !strings.HasPrefix(string(formatted), opening) {
			continue
		}

		all[this].fix = strings.TrimPrefix(string(formatted), opening)
		for _, inner := range inside {
			all[inner].fix = ""
		}
	}

	slices.Reverse(all)

	return all
}
