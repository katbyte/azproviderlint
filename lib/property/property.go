// Package property says what a schema property is wherever its schema is written: in place, or handed back by a function in whichever package
// that is, so a check reads a property from a shared helper as it would one written out.
package property

import (
	"go/ast"
	"go/constant"
	"go/types"
	"reflect"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/facts"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reads the properties in a package once, for every check that needs them. It records what each function that hands back a property
// or a block gives, so a check can read one that comes from a function where the function is used.
var Analyzer = &analysis.Analyzer{
	Name:       "property",
	Doc:        "find what each schema property says, following schemas handed back by functions",
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	FactTypes:  []analysis.Fact{(*gives)(nil), (*holds)(nil)},
	ResultType: reflect.TypeFor[*Found](),
}

// run refers to Analyzer, so it is wired up here and not in the literal, which would be an initialization cycle.
func init() { Analyzer.Run = run }

// Schema is what a property's schema says. The plugin sdk and the framework both keep their schema types in a package named schema.
type Schema struct {
	// Type is the plugin SDK's type, such as TypeString, or the framework attribute's, such as BoolAttribute.
	Type string

	Required, Optional, Computed, WriteOnly, Deprecated, Framework bool

	// Validated is whether it has validation, OnlyNotEmpty whether that only says a string is not empty, and Validate the ValidateFunc as written.
	Validated, OnlyNotEmpty bool
	Validate                string

	// Default is whether it has a Default or DefaultFunc, OneOf whether it has an AtLeastOneOf or ExactlyOneOf, and Inputs whether it sets a
	// field that only matters for a value the user sets, one of InputOnly.
	Default, OneOf, Inputs bool

	// Elem is what a list, set or map holds: Resource for blocks, otherwise the Type of its values, or "" when that cannot be read. ElemValidated
	// and ElemOnlyNotEmpty say of its values what Validated and OnlyNotEmpty say of a property.
	Elem                            string
	ElemValidated, ElemOnlyNotEmpty bool
}

// Set is whether the user sets the property.
func (s Schema) Set() bool { return s.Required || s.Optional }

// Values is what the values a list, set or map holds say, as far as the property says.
func (s Schema) Values() Schema {
	return Schema{Type: s.Elem, Validated: s.ElemValidated, OnlyNotEmpty: s.ElemOnlyNotEmpty, Framework: s.Framework, Deprecated: s.Deprecated}
}

// ReadOnly is whether the property is only read back.
func (s Schema) ReadOnly() bool { return s.Computed && !s.Set() }

// Block is what the properties of a block say, for one whose properties come from a function.
type Block struct {
	// Properties is how many there are, and Known whether what each one says could be read.
	Properties int
	Known      bool

	// Stops is whether one keeps the block from being empty: it is Required, has a Default, or has an AtLeastOneOf or ExactlyOneOf. Settable is
	// whether one is Required or Optional, or sets one of InputOnly.
	Stops, Settable bool
}

// InputOnly are the fields that only matter for a value the user sets: what it is checked, defaulted, normalised, counted or constrained by. It is
// the plugin SDK's own list for a property that is only read back, plus RequiredWith and WriteOnly.
var InputOnly = []string{"Default", "DefaultFunc", "InputDefault", "ValidateFunc", "ValidateDiagFunc", "DiffSuppressFunc", "DiffSuppressOnRefresh", "StateFunc", "MaxItems", "MinItems", "AtLeastOneOf", "ExactlyOneOf", "ConflictsWith", "RequiredWith", "WriteOnly"}

// IsSet reports whether e gives a field a value: not nil, and not a zero constant.
func IsSet(pass *analysis.Pass, e ast.Expr) bool {
	tv := pass.TypesInfo.Types[e]
	if e == nil || tv.IsNil() {
		return false
	}
	if tv.Value == nil {
		return true
	}

	switch tv.Value.Kind() {
	case constant.Bool:
		return constant.BoolVal(tv.Value)

	case constant.String:
		return constant.StringVal(tv.Value) != ""

	case constant.Int, constant.Float:
		return constant.Sign(tv.Value) != 0

	case constant.Complex, constant.Unknown:
	}

	return true
}

// Kind is what a value holds.
type Kind uint8

const (
	String Kind = iota + 1
	Int
	Float
)

// kinds are what each plugin SDK type and framework attribute holds.
var kinds = map[string]Kind{"TypeString": String, "TypeInt": Int, "TypeFloat": Float, "StringAttribute": String, "Int64Attribute": Int, "Int32Attribute": Int, "Float64Attribute": Float, "Float32Attribute": Float, "NumberAttribute": Float}

// Value is one value the user sets, a property or the values of a list or set, for a check of its validation.
type Value struct {
	// At is where it is reported: the key naming its property, the schema itself, or the call it comes from.
	At ast.Node

	// Name is the key naming its property, when there is one.
	Name ast.Expr

	Kind Kind

	// Element is whether it is the values of a list or set.
	Element bool

	// Checked is whether it has validation, and OnlyNotEmpty whether that only says a string is not empty.
	Checked, OnlyNotEmpty bool

	// Framework is whether it is a framework attribute, which is validated with Validators and not a ValidateFunc.
	Framework bool

	// From is the function from another module, such as go-azure-helpers, it comes from. It is nil when the schema is written in the provider.
	From *types.Func
}

// Found is what Analyzer reads in a package.
type Found struct {
	module  string
	schemas map[*ast.CallExpr]Schema
	blocks  map[*ast.CallExpr]Block
	values  []Value
}

func found(pass *analysis.Pass) *Found {
	if f, ok := pass.ResultOf[Analyzer].(*Found); ok {
		return f
	}

	return &Found{}
}

// called is the call an expression makes, past any & or brackets.
func called(e ast.Expr) *ast.CallExpr {
	if ref, isRef := ast.Unparen(e).(*ast.UnaryExpr); isRef {
		e = ref.X
	}
	call, _ := ast.Unparen(e).(*ast.CallExpr)

	return call
}

// Of returns what the property a call hands back says, for a check that requires Analyzer. It is false when e is not a call to a function that
// hands back a property whose schema could be read.
func Of(pass *analysis.Pass, e ast.Expr) (Schema, bool) {
	s, ok := found(pass).schemas[called(e)]

	return s, ok
}

// BlockOf returns what the properties a call hands back say, as a block's Elem or Schema, for a check that requires Analyzer.
func BlockOf(pass *analysis.Pass, e ast.Expr) (Block, bool) {
	b, ok := found(pass).blocks[called(e)]

	return b, ok
}

// Outside reports whether fn is in another module than the package being checked, such as go-azure-helpers, which the provider cannot change.
func Outside(pass *analysis.Pass, fn *types.Func) bool { return found(pass).outside(fn) }

func (f *Found) outside(fn *types.Func) bool {
	return fn != nil && fn.Pkg() != nil && f.module != "" && fn.Pkg().Path() != f.module && !strings.HasPrefix(fn.Pkg().Path(), f.module+"/")
}

// Values returns every value the user sets in the package, for a check of its validation that requires Analyzer.
func Values(pass *analysis.Pass) []Value { return found(pass).values }

// gives is the fact on a function that hands back one property.
type gives struct{ S Schema }

func (*gives) AFact() {}

func (f *gives) String() string {
	said := []string{f.S.Type}
	for i, is := range []bool{f.S.Required, f.S.Optional, f.S.Computed, f.S.WriteOnly, f.S.Deprecated, f.S.Framework, f.S.Validated, f.S.OnlyNotEmpty, f.S.Default, f.S.OneOf, f.S.Inputs, f.S.ElemValidated, f.S.ElemOnlyNotEmpty} {
		if is {
			said = append(said, []string{"required", "optional", "computed", "write-only", "deprecated", "framework", "validated", "not-empty", "default", "one-of", "inputs", "elem-validated", "elem-not-empty"}[i])
		}
	}
	if f.S.Elem != "" {
		said = append(said, "elem="+f.S.Elem)
	}

	return "gives(" + strings.Join(said, ",") + ")"
}

// holds is the fact on a function that hands back a block's properties, as a Resource or a map of them.
type holds struct{ B Block }

func (*holds) AFact() {}

func (f *holds) String() string {
	said := []string{strings.Repeat("|", f.B.Properties)}
	for i, is := range []bool{f.B.Known, f.B.Stops, f.B.Settable} {
		if is {
			said = append(said, []string{"known", "stops", "settable"}[i])
		}
	}

	return "holds(" + strings.Join(said, ",") + ")"
}

// Resource is the plugin SDK's type for a block, whose Schema holds its properties.
const Resource = "Resource"

// notEmpty are the checks that say only that a string has something in it.
var notEmpty = []string{"StringIsNotEmpty", "StringIsNotWhiteSpace", "NoZeroValues"}

// element is what a list or set says about the schema of the values in it: the name of the property, and whether the user sets it.
type element struct {
	of  ast.Expr
	set bool
}

// schemaType is the name of a value's type when it is one of the plugin sdk's or the framework's, which both keep them in a package named schema.
func schemaType(t types.Type) string {
	if t == nil {
		return ""
	}
	if ptr, isPointer := types.Unalias(t).(*types.Pointer); isPointer {
		t = ptr.Elem()
	}

	named, isNamed := types.Unalias(t).(*types.Named)
	if !isNamed || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" {
		return ""
	}

	return named.Obj().Name()
}

func run(pass *analysis.Pass) (any, error) {
	// the values each function hands back, outside test files
	returns := func(fn *types.Func, decl *ast.FuncDecl) ([]ast.Expr, bool) {
		if fn.Signature().Results().Len() != 1 || strings.HasSuffix(pass.Fset.Position(decl.Pos()).Filename, "_test.go") {
			return nil, false
		}

		var returned []ast.Expr
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			if ret, isReturn := n.(*ast.ReturnStmt); isReturn && len(ret.Results) == 1 {
				returned = append(returned, ret.Results[0])
			}
			_, isFunc := n.(*ast.FuncLit)

			return !isFunc
		})

		return returned, len(returned) > 0
	}

	// judge reads one property's schema, with what each function hands back known so far
	var judge func(s *facts.Set[gives, *gives], lit *ast.CompositeLit) (Schema, map[string]ast.Expr, bool)
	judge = func(s *facts.Set[gives, *gives], lit *ast.CompositeLit) (Schema, map[string]ast.Expr, bool) {
		var out Schema
		name := schemaType(pass.TypesInfo.TypeOf(lit))
		if name == "" || name == Resource {
			return out, nil, false
		}

		fields := map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			if kv, isPair := elt.(*ast.KeyValueExpr); isPair {
				if key, isField := kv.Key.(*ast.Ident); isField {
					fields[key.Name] = kv.Value
				}
			}
		}

		out.Required, out.Optional, out.Computed, out.WriteOnly = astx.IsTrueConstant(pass, fields["Required"]), astx.IsTrueConstant(pass, fields["Optional"]), astx.IsTrueConstant(pass, fields["Computed"]), astx.IsTrueConstant(pass, fields["WriteOnly"])
		out.Deprecated = fields["Deprecated"] != nil || fields["DeprecationMessage"] != nil
		out.Default, out.OneOf = IsSet(pass, fields["Default"]) || IsSet(pass, fields["DefaultFunc"]), IsSet(pass, fields["AtLeastOneOf"]) || IsSet(pass, fields["ExactlyOneOf"])
		for _, input := range InputOnly {
			out.Inputs = out.Inputs || IsSet(pass, fields[input])
		}

		// a framework attribute says what it holds in its type, and is validated with Validators
		if name != "Schema" {
			list, isList := fields["Validators"].(*ast.CompositeLit)
			out.Type, out.Framework = name, true
			out.Validated = fields["Validators"] != nil && (!isList || len(list.Elts) > 0)
			out.OnlyNotEmpty = out.Validated && kinds[name] == String && isList && len(list.Elts) == 1 && strings.HasSuffix(types.ExprString(list.Elts[0]), "LengthAtLeast(1)")

			return out, fields, true
		}

		// a framework resource's own schema has the same name as a plugin SDK property's, and no Type
		if fields["Type"] == nil {
			return out, fields, false
		}

		typed := types.ExprString(fields["Type"])
		out.Type, out.Validate = typed[strings.LastIndex(typed, ".")+1:], types.ExprString(fields["ValidateFunc"])
		if fields["ValidateFunc"] == nil {
			out.Validate = ""
		}
		out.Validated = (fields["ValidateFunc"] != nil && !astx.IsNilValue(pass, fields["ValidateFunc"])) || (fields["ValidateDiagFunc"] != nil && !astx.IsNilValue(pass, fields["ValidateDiagFunc"]))
		out.OnlyNotEmpty = out.Validated && kinds[out.Type] == String && fields["ValidateDiagFunc"] == nil && slices.Contains(notEmpty, out.Validate[strings.LastIndex(out.Validate, ".")+1:])

		// what a list, set or map holds, written out or from a function
		elem := fields["Elem"]
		if ref, isRef := elem.(*ast.UnaryExpr); isRef {
			elem = ref.X
		}
		switch elem := elem.(type) {
		case *ast.CompositeLit:
			if schemaType(pass.TypesInfo.TypeOf(elem)) == Resource {
				out.Elem = Resource
			} else if inner, _, isValue := judge(s, elem); isValue {
				out.Elem, out.ElemValidated, out.ElemOnlyNotEmpty = inner.Type, inner.Validated, inner.OnlyNotEmpty
			}

		case *ast.CallExpr:
			if schemaType(pass.TypesInfo.TypeOf(elem)) == Resource {
				out.Elem = Resource
			} else if inner := s.Get(astx.CalledFunc(pass, elem)).S; inner.Type != "" {
				out.Elem, out.ElemValidated, out.ElemOnlyNotEmpty = inner.Type, inner.Validated, inner.OnlyNotEmpty
			}
		}

		return out, fields, true
	}

	// what each function that hands back one property gives: the first value it returns whose schema can be read
	schemas, err := facts.Run(pass, Analyzer, returns, func(s *facts.Set[gives, *gives], returned []ast.Expr) gives {
		for _, value := range returned {
			if call := called(value); call != nil {
				if g := s.Get(astx.CalledFunc(pass, call)); g.S.Type != "" {
					return g
				}
			}

			value = ast.Unparen(value)
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}
			if lit, isLit := value.(*ast.CompositeLit); isLit {
				if out, _, isSchema := judge(s, lit); isSchema {
					return gives{S: out}
				}
			}
		}

		return gives{}
	})
	if err != nil {
		return nil, err
	}

	// summarise says what the properties a block holds say: written out as a map, a Resource whose Schema is one, or from a function
	var summarise func(b *facts.Set[holds, *holds], e ast.Expr) (Block, bool)
	summarise = func(b *facts.Set[holds, *holds], e ast.Expr) (Block, bool) {
		if call := called(e); call != nil {
			got := b.Get(astx.CalledFunc(pass, call)).B

			return got, got != Block{}
		}

		e = ast.Unparen(e)
		if ref, isRef := e.(*ast.UnaryExpr); isRef {
			e = ref.X
		}
		lit, isLit := e.(*ast.CompositeLit)
		if !isLit {
			return Block{}, false
		}

		if schemaType(pass.TypesInfo.TypeOf(lit)) == Resource {
			for _, elt := range lit.Elts {
				if kv, isPair := elt.(*ast.KeyValueExpr); isPair {
					if key, isField := kv.Key.(*ast.Ident); isField && key.Name == "Schema" {
						return summarise(b, kv.Value)
					}
				}
			}

			return Block{}, false
		}

		if _, isMap := pass.TypesInfo.TypeOf(lit).Underlying().(*types.Map); !isMap {
			return Block{}, false
		}

		out := Block{Known: true}
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := ast.Unparen(kv.Value)
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}

			var property Schema
			var isSchema bool
			switch value := value.(type) {
			case *ast.CompositeLit:
				property, _, isSchema = judge(schemas, value)

			case *ast.CallExpr:
				property = schemas.Get(astx.CalledFunc(pass, value)).S
				isSchema = property.Type != ""
			}

			out.Properties++
			out.Known = out.Known && isSchema
			out.Stops = out.Stops || property.Required || property.Default || property.OneOf
			out.Settable = out.Settable || property.Set() || property.Inputs
		}

		return out, true
	}

	// what each function that hands back a block's properties holds
	blocks, err := facts.Run(pass, Analyzer, returns, func(b *facts.Set[holds, *holds], returned []ast.Expr) holds {
		for _, value := range returned {
			if out, isBlock := summarise(b, value); isBlock {
				return holds{B: out}
			}
		}

		return holds{}
	})
	if err != nil {
		return nil, err
	}

	out := &Found{schemas: map[*ast.CallExpr]Schema{}, blocks: map[*ast.CallExpr]Block{}}

	// a function in the provider's own code is checked where its schema is written; one from another module, such as go-azure-helpers, where it
	// is used, as the provider cannot change it
	parts := strings.Split(pass.Pkg.Path(), "/")
	out.module = parts[0]
	if strings.Contains(parts[0], ".") && len(parts) >= 3 {
		out.module = strings.Join(parts[:3], "/")
	}
	if pass.Module != nil && pass.Module.Path != "" {
		out.module = pass.Module.Path
	}

	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return out, nil
	}

	// every call in the package to a function that hands back a property or a block, for a check that meets it
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, isCall := n.(*ast.CallExpr)
		if !isCall {
			return
		}

		fn := astx.CalledFunc(pass, call)
		if g := schemas.Get(fn); g.S.Type != "" {
			out.schemas[call] = g.S
		}
		if h := blocks.Get(fn); h.B != (Block{}) {
			out.blocks[call] = h.B
		}
	})

	// a state migration's schema is what old state looked like, which nobody sets
	if tf.InStateMigration(pass) {
		return out, nil
	}

	// the values the user sets: a property is named by its key in the map it sits in, and a list is met before the schema of its values is
	keys, elements := map[*ast.CompositeLit]ast.Expr{}, map[*ast.CompositeLit]element{}
	add := func(at ast.Node, name ast.Expr, property Schema, isElement bool, from *types.Func) {
		if kinds[property.Type] != 0 && !property.Deprecated {
			out.values = append(out.values, Value{At: at, Name: name, Kind: kinds[property.Type], Element: isElement, Checked: property.Validated, OnlyNotEmpty: property.OnlyNotEmpty, Framework: property.Framework, From: from})
		}
	}
	listed := func(property Schema) bool {
		return property.Set() && (property.Type == "TypeList" || property.Type == "TypeSet")
	}

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		// a schema made up inside a test is never shipped
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return
		}

		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := ast.Unparen(kv.Value)
			if ref, isRef := value.(*ast.UnaryExpr); isRef {
				value = ref.X
			}
			if inner, isInner := value.(*ast.CompositeLit); isInner {
				keys[inner] = kv.Key
			}

			// a property whose whole schema comes from a function outside the provider
			call := called(kv.Value)
			if _, isField := kv.Key.(*ast.Ident); isField || call == nil || !out.outside(astx.CalledFunc(pass, call)) {
				continue
			}
			if property, isProperty := out.schemas[call]; isProperty && property.Set() {
				add(kv.Key, kv.Key, property, false, astx.CalledFunc(pass, call))
				if listed(property) {
					add(kv.Key, kv.Key, property.Values(), true, astx.CalledFunc(pass, call))
				}
			}
		}

		property, fields, isSchema := judge(schemas, lit)
		if !isSchema {
			return
		}

		name, set := keys[lit], property.Set()
		of, isElement := elements[lit]
		if isElement {
			name, set = of.of, of.set
		}

		elem := fields["Elem"]
		if ref, isRef := elem.(*ast.UnaryExpr); isRef {
			elem = ref.X
		}
		switch elem := elem.(type) {
		case *ast.CompositeLit:
			elements[elem] = element{of: keys[lit], set: listed(property)}

		case *ast.CallExpr:
			// the values' schema from a function outside the provider
			if listed(property) && out.outside(astx.CalledFunc(pass, elem)) {
				add(elem, keys[lit], property.Values(), true, astx.CalledFunc(pass, elem))
			}
		}

		at := ast.Node(lit)
		if keys[lit] != nil {
			at = keys[lit]
		}
		if set {
			add(at, name, property, isElement, nil)
		}
	})

	return out, nil
}
