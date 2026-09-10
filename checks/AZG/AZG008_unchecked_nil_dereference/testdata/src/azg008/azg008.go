package azg008

import (
	"context"
	"os"

	"github.com/example/helpers"
	"github.com/example/sdk/widgets"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
)

type Status string

type properties struct {
	Status *Status
	Count  *int
	Name   *string
}

type model struct {
	Properties *properties
}

type data struct{}

func (data) Set(string, interface{}) {}

func use(interface{}) {}

// Should be flagged: bare dereference of an optional field, converted to string — the fix is
// pointer.FromEnum.
func invalidEnumConversion(d data, props properties) {
	d.Set("status", string(*props.Status)) // want "dereference of possibly-nil `props.Status` may panic - add a nil check or use pointer.From"
}

// Should be flagged: bare dereference of a non-enum field — the fix is pointer.From.
func invalidPlainDeref(d data, props properties) {
	d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
}

// Should be flagged: a deeper selector chain with no guard on the dereferenced link.
func invalidChain(d data, m model) {
	d.Set("name", *m.Properties.Name) // want "dereference of possibly-nil `m.Properties.Name` may panic - add a nil check or use pointer.From"
}

// Should be flagged: a guard on a prefix of the chain does not cover the dereferenced field.
func invalidPrefixGuardOnly(d data, m model) {
	if m.Properties != nil {
		d.Set("name", *m.Properties.Name) // want "dereference of possibly-nil `m.Properties.Name` may panic - add a nil check or use pointer.From"
	}
}

// Should NOT be flagged here: an assignment target needs the pointer itself — AZG009's.
func writeTargetIsAZG009(props properties) {
	*props.Count = 1
}

// Should NOT be flagged: enclosing if proves the field non-nil.
func validIfGuard(d data, props properties) {
	if props.Status != nil {
		d.Set("status", string(*props.Status))
	}
}

// Should NOT be flagged: the && left operand proves the field non-nil for the right.
func validShortCircuitAnd(props properties) bool {
	return props.Count != nil && *props.Count > 0
}

// Should NOT be flagged: the || left operand short-circuits when the field is nil.
func validShortCircuitOr(props properties) bool {
	return props.Count == nil || *props.Count == 0
}

// Should NOT be flagged: the else branch of a pure-|| nil condition.
func validElseBranch(d data, props properties) {
	if props.Status == nil || *props.Status == "" {
		d.Set("status", "")
	} else {
		d.Set("status", string(*props.Status))
	}
}

// Should NOT be flagged: an early return guards everything below it.
func validEarlyReturn(d data, props properties) {
	if props.Status == nil {
		return
	}
	d.Set("status", string(*props.Status))
}

// Should NOT be flagged: an early exit via os.Exit or panic also terminates.
func validEarlyExit(props properties) int {
	if props.Count == nil {
		os.Exit(1)
	}
	return *props.Count
}

// Should NOT be flagged: a pure-|| early return covers both fields.
func validOrEarlyReturn(props properties) int {
	if props.Count == nil || props.Name == nil {
		return 0
	}
	use(*props.Name)
	return *props.Count
}

// Should NOT be flagged: the variable only ever holds provably non-nil values.
func validNonNilSources() (int, string) {
	count := new(int)
	name := pointer.To("x")
	status := new(Status)
	_ = status
	return *count, *name
}

// Should be flagged: reassignment to an unknown value invalidates the earlier guard.
func invalidReassigned(props properties, next func() *int) int {
	if props.Count == nil {
		return 0
	}
	count := next()
	return *count // want "dereference of possibly-nil `count` may panic - add a nil check or use pointer.From"
}

// Should be flagged: reassignment inside the guarded body invalidates the enclosing guard.
func invalidReassignedInsideGuard(next func() *int) int {
	x := next()
	if x != nil {
		x = next()
		return *x // want "dereference of possibly-nil `x` may panic - add a nil check or use pointer.From"
	}
	return 0
}

// Should be flagged: reassignment in a nested block invalidates the earlier early return.
func invalidReassignedNestedBlock(props properties, next func() *int) int {
	if props.Count == nil {
		return 0
	}
	{
		props.Count = next()
		return *props.Count // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
	}
}

// Should NOT be flagged: the reassignment inside the guard is itself provably non-nil.
func validReassignedNonNil(next func() *int) int {
	x := next()
	if x != nil {
		x = new(int)
		return *x
	}
	return 0
}

// Should NOT be flagged: a switch case condition proves the field non-nil.
func validSwitchCase(props properties) int {
	switch {
	case props.Count != nil:
		return *props.Count
	}
	return 0
}

// Should NOT be flagged: a guard outside a closure still covers dereferences inside it.
func validGuardOutsideClosure(d data, props properties) {
	if props.Status != nil {
		f := func() { d.Set("status", string(*props.Status)) }
		f()
	}
}

// Should NOT be flagged: the alias's source chain is guarded by the enclosing if.
func validAliasedGuard(d data, props properties) {
	s := props.Status
	if props.Status != nil {
		d.Set("status", string(*s))
	}
}

// Should NOT be flagged: the alias itself is guarded.
func validGuardedAliasDirect(d data, props properties) {
	s := props.Status
	if s != nil {
		d.Set("status", string(*s))
	}
}

// Should NOT be flagged: for-loop condition proves non-nil inside the body.
type node struct{ next *node }

func validForCond(n *node) int {
	count := 0
	for n != nil {
		n = (*n).next
		count++
	}
	return count
}

// Should NOT be flagged: `x, err := f()` followed by an error exit — the Go contract makes
// the other results valid.
func validErrContract(parse func(string) (*properties, error)) int {
	props, err := parse("x")
	if err != nil {
		return 0
	}
	return len((*props).Name2())
}

// Should be flagged: the error is discarded, so nothing proves the pointer valid.
func invalidDiscardedErr(parse func(string) (*properties, error)) int {
	props, _ := parse("x")
	return len((*props).Name2()) // want "dereference of possibly-nil `props` may panic - add a nil check or use pointer.From"
}

func (properties) Name2() string { return "" }

// Should NOT be flagged: the alias's source chain was nil-checked before the assignment.
func validAliasOfGuardedChain(d data, m model) {
	if m.Properties == nil {
		return
	}
	payload := m.Properties
	d.Set("status", string((*payload).StatusString()))
}

// Should be flagged: only the alias's source prefix was checked, not the derefd field.
func invalidAliasFieldUnguarded(d data, m model) {
	if m.Properties == nil {
		return
	}
	payload := m.Properties
	d.Set("count", *payload.Count) // want "dereference of possibly-nil `payload.Count` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: `x, ok := f()` followed by a !ok exit — the comma-ok contract.
func validOkContract(lookup func(string) (*int, bool)) int {
	suffix, ok := lookup("x")
	if !ok {
		return 0
	}
	return *suffix
}

func (properties) StatusString() string { return "" }

// Should NOT be flagged: a bare pointer parameter's nil contract belongs to its callers.
func validBareParam(id *int) int {
	return *id
}

// Should be flagged: a field dereference through a parameter is always in scope.
func invalidParamField(props *properties) int {
	return *props.Count // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: the if-init alias is nil-checked, which covers its source chain.
func validIfInitAliasGuard(d data, props properties) {
	if days := props.Count; days != nil {
		d.Set("count", int(*props.Count))
	}
}

// Should NOT be flagged: an alias declared earlier in the block is nil-checked.
func validBlockAliasGuard(d data, props properties) {
	days := props.Count
	if days != nil {
		d.Set("count", int(*props.Count))
	}
}

// Should NOT be flagged: an alias's early exit covers its source chain.
func validAliasEarlyExit(d data, props properties) {
	days := props.Count
	if days == nil {
		return
	}
	d.Set("count", int(*props.Count))
}

// Should NOT be flagged: a prefix alias carries the remaining path.
func validPrefixAliasGuard(d data, m model) {
	v := m
	if v.Properties != nil {
		d.Set("props", *m.Properties)
	}
}

// Should be flagged: the source chain was reassigned after the alias was taken.
func invalidStaleAlias(d data, props properties, other *int) {
	days := props.Count
	props.Count = other
	if days != nil {
		d.Set("count", int(*props.Count)) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
	}
}

// Should be flagged: the alias was reassigned before the nil check.
func invalidReassignedAlias(d data, props properties, other *int) {
	days := props.Count
	days = other
	if days != nil {
		d.Set("count", int(*props.Count)) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
	}
}

// Should NOT be flagged: these contexts need an addressable pointee, so pointer.From cannot
// stand in — assignment through a field, an array index or slice, or a pointer-receiver
// method; they are AZG009's to report.
type counter struct{ n int }

func (c *counter) inc() { c.n++ }

type holder struct {
	C   *counter
	Arr *[2]int
}

func validNeedsAddressable(h holder) {
	(*h.C).n = 1
	(*h.C).inc()
	(*h.Arr)[0] = 1
	use((*h.Arr)[:])
	use(&(*h.C).n)
}

// Should NOT be flagged either: writing through a slice or map index after pointer.From
// still panics (nil map, empty slice), so the rewrite would fix nothing.
func validIndexWrite(b bag) {
	(*b.M)["k"] = 1
	(*b.S)[0] = 1
}

// Should be flagged: a value-receiver method or a read through the deref'd value copies fine.
type bag struct {
	M *map[string]int
	S *[]int
}

func (c counter) get() int { return c.n }

func invalidCopyable(h holder, b bag) {
	use((*h.C).get()) // want "dereference of possibly-nil `h.C` may panic - add a nil check or use pointer.From"
	use((*h.Arr)[0])  // want "dereference of possibly-nil `h.Arr` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: an explicitly instantiated pointer.To* call always allocates.
func validGenericToEnum(d data, props properties) {
	props.Status = pointer.ToEnum[Status]("x")
	d.Set("status", string(*props.Status))
}

type resp struct{ Model *model }

// Should NOT be flagged: `if x == nil { x = <non-nil> }` default-init.
func validDefaultInit(d data, props properties) {
	if props.Count == nil {
		props.Count = pointer.To(0)
	}
	d.Set("count", *props.Count)
}

// Should be flagged: the default-init source is not provably non-nil.
func invalidDefaultInitUnknown(d data, props properties, other *int) {
	if props.Count == nil {
		props.Count = other
	}
	d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: the err companion is checked in the if the call is the init of.
func validIfInitErr(parse func(string) (*properties, error)) int {
	var p *properties
	var err error
	if p, err = parse("x"); err != nil {
		return 0
	}
	return len((*p).Name2())
}

// Should NOT be flagged: `if x, ok := f(); ok { *x }`.
func validIfInitOk(lookup func(string) (*int, bool)) int {
	if v, ok := lookup("x"); ok {
		return *v
	}
	return 0
}

// Should NOT be flagged: the body of `if err == nil` and the else of `if err != nil`.
func validErrNilBody(d data, parse func(string) (*properties, error)) {
	p, err := parse("x")
	if err == nil {
		d.Set("name", (*p).Name2())
	}
	q, err := parse("y")
	if err != nil {
		d.Set("name", "")
	} else {
		d.Set("name", (*q).Name2())
	}
}

// Should NOT be flagged: aliases declared in an enclosing if's init, in either direction.
func validEnclosingInitAlias(d data, r resp) {
	if m := r.Model; m != nil {
		if m.Properties != nil {
			d.Set("props", *r.Model.Properties)
		}
	}
	if m := r.Model; r.Model != nil {
		d.Set("model", *m)
	}
}

// Should NOT be flagged: commonids composite IDs always populate First and Second.
func validCompositeID(d data) {
	id, err := commonids.ParseCompositeResourceID("x", &properties{}, &model{})
	if err != nil {
		return
	}
	d.Set("first", *id.First)
	built := commonids.NewCompositeResourceID(&properties{}, &model{})
	d.Set("second", *built.Second)
}

// Should NOT be flagged: a non-zero pointer.From result proves the pointer non-nil.
func validFromNonZero(d data, props properties) {
	if pointer.From(props.Name) != "" {
		d.Set("name", *props.Name)
	}
	if len(pointer.From(props.Name)) > 0 {
		d.Set("name", *props.Name)
	}
	if 0 < len(pointer.From(props.Name)) {
		d.Set("name", *props.Name)
	}
}

// Should NOT be flagged: code after a terminating `if x == nil` runs only when x != nil, else
// branch or not.
func validEarlyExitWithElse(d data, props properties) {
	if props.Count == nil {
		return
	} else {
		d.Set("has", true)
	}
	d.Set("count", *props.Count)
}

// Should be flagged: the err companion was overwritten before the check.
func invalidErrReassigned(d data, parse func(string) (*properties, error), other func() error) {
	p, err := parse("x")
	err = other()
	if err == nil {
		d.Set("name", (*p).Name2()) // want "dereference of possibly-nil `p` may panic - add a nil check or use pointer.From"
	}
	q, err := parse("y")
	err = other()
	if err != nil {
		return
	}
	d.Set("name", (*q).Name2()) // want "dereference of possibly-nil `q` may panic - add a nil check or use pointer.From"
}

// Should be flagged: the alias's source chain was reassigned after the copy.
func invalidForwardAliasStale(d data, r resp, other *model) {
	m := r.Model
	r.Model = other
	if r.Model != nil {
		d.Set("model", *m) // want "dereference of possibly-nil `m` may panic - add a nil check or use pointer.From"
	}
}

// Should be flagged: the else branch reassigns the guarded field.
func invalidEarlyExitElseReassign(d data, props properties, other *int) {
	if props.Count == nil {
		return
	} else {
		props.Count = other
	}
	d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
}

// Should be flagged: an inner if's init reassigns what the outer if guarded.
func invalidIfInitReassign(d data, props properties, other *int, flag bool) {
	if props.Count != nil {
		if props.Count = other; flag {
			d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
		}
	}
}

// Should be flagged: a type assertion's ok proves the dynamic type, not that the pointer is
// non-nil.
func invalidTypeAssertOk(d data, v interface{}) {
	if p, ok := v.(*int); ok {
		d.Set("count", *p) // want "dereference of possibly-nil `p` may panic - add a nil check or use pointer.From"
	}
}

// Should NOT be flagged / should be flagged: a composite ID's field is as nil as the argument
// that built it.
func compositeIDArgs(d data, parsed func() (*properties, error), other *model) {
	first, err := parsed()
	if err != nil {
		return
	}
	id := commonids.NewCompositeResourceID(first, other)
	d.Set("first", *id.First)
	d.Set("second", *id.Second) // want "dereference of possibly-nil `id.Second` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: the field was set from a non-nil source in the literal that built
// the struct, directly or through a nested &literal.
func validLiteralField(d data, r resp) {
	m := model{Properties: &properties{Count: pointer.To(1)}}
	d.Set("count", *m.Properties.Count)
	d.Set("props", *m.Properties)
	pr := &properties{Name: pointer.To("x")}
	d.Set("name", *pr.Name)
	if r.Model != nil {
		wrapped := resp{Model: r.Model}
		d.Set("model", *wrapped.Model)
	}
}

// Should be flagged: the literal left the field zero, aliased an unguarded value, or the
// field was reassigned afterwards.
func invalidLiteralField(d data, r resp, other *int) {
	m := model{Properties: &properties{Name: pointer.To("x")}}
	d.Set("count", *m.Properties.Count) // want "dereference of possibly-nil `m.Properties.Count` may panic - add a nil check or use pointer.From"
	wrapped := resp{Model: r.Model}
	d.Set("model", *wrapped.Model) // want "dereference of possibly-nil `wrapped.Model` may panic - add a nil check or use pointer.From"
	pr := properties{Count: pointer.To(1)}
	pr.Count = other
	d.Set("count", *pr.Count) // want "dereference of possibly-nil `pr.Count` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: an immediately-invoked func literal returning a non-nil source on
// every path.
func validIIFE(d data, flag bool) {
	pr := properties{Status: func() *Status {
		if flag {
			return pointer.ToEnum[Status]("a")
		}
		return pointer.ToEnum[Status]("b")
	}()}
	d.Set("status", string(*pr.Status))
}

// Should be flagged: one path of the func literal returns nil.
func invalidIIFE(d data, flag bool) {
	pr := properties{Status: func() *Status {
		if flag {
			return nil
		}
		return pointer.ToEnum[Status]("b")
	}()}
	d.Set("status", string(*pr.Status)) // want "dereference of possibly-nil `pr.Status` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: helpers proven never to return nil (via facts from their package, and
// from this one), whether dereferenced directly, through a field, or bound alongside an
// unchecked error.
func expandLocal() *properties { return &properties{} }

func validHelperNonNil(d data, raw []interface{}) {
	direct := helpers.ExpandStringSlice(raw)
	d.Set("direct", *direct)
	pr := properties{Name: pointer.To("x")}
	m := model{Properties: expandLocal()}
	d.Set("props", *m.Properties)
	d.Set("name", *pr.Name)
	n, _ := helpers.Parse("abc")
	d.Set("n", *n)
}

// Should be flagged: the helper returns nil on some path.
func invalidHelperMaybe(d data, raw []interface{}) {
	v := helpers.ExpandMaybe(raw)
	d.Set("maybe", *v) // want "dereference of possibly-nil `v` may panic - add a nil check or use pointer.From"
}

type widgetHolder struct{ Model *widgets.Widget }

// Should be flagged WITH a fix: a GET sends no body, whatever the argument.
func invalidNotPayload(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	_ = c.Get(ctx, "id", *h.Model) // want "may panic - add a nil check or use pointer.From"
}

// Should be flagged WITH a fix: the copy never reaches a write.
func invalidCopyNotSent(d data, h widgetHolder) {
	env := widgets.Envelope{Widget: *h.Model} // want "may panic - add a nil check or use pointer.From"
	d.Set("env", env)
}

// Should be flagged WITHOUT a fix: the value is marshalled as a PUT body by the callee (known
// through requestbody facts, including via a delegating wrapper), passed directly or through
// a local.
func invalidPayloadDirect(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	_ = c.CreateOrUpdate(ctx, "id", *h.Model) // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
}

func invalidPayloadWrapper(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	_ = c.ForRegionCreateOrUpdateThenPoll(ctx, "id", *h.Model) // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
}

func invalidPayloadViaLocal(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	m := *h.Model // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
	m.Name = nil
	_ = c.CreateOrUpdate(ctx, "id", m)
}

// Should be flagged WITHOUT a fix: the value is copied into a payload that is sent — through a
// literal field, a field assignment, or by address.
func invalidPayloadLiteralField(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	env := widgets.Envelope{Widget: *h.Model} // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
	_ = c.Put(ctx, "id", env)
}

func invalidPayloadFieldAssign(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	env := widgets.Envelope{}
	env.Widget = *h.Model // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
	_ = c.Put(ctx, "id", env)
}

func invalidPayloadAddrOf(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	m := *h.Model // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
	_ = c.PutPtr(ctx, "id", &m)
}

// Should be flagged: a write nested in a block, a loop, or an if-init cancels the guard;
// only the last write of a default-init counts; alias staleness follows the derived path.
func invalidNestedReassign(d data, props properties, other *int, flag bool) {
	if props.Count == nil {
		return
	}
	if flag {
		props.Count = other
	}
	d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
}

func invalidLoopBackEdge(d data, props properties, next func() *int) {
	if props.Count != nil {
		for i := 0; i < 3; i++ {
			d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
			props.Count = next()
		}
	}
}

func invalidInitInCond(d data, props properties, other *int) {
	if props.Count != nil {
		if props.Count = other; *props.Count > 0 { // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
			d.Set("positive", true)
		}
	}
}

func invalidDefaultInitOverwritten(d data, props properties, other *int) {
	if props.Count == nil {
		props.Count = pointer.To(0)
		props.Count = other
	}
	d.Set("count", *props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check or use pointer.From"
}

func invalidAliasFieldStale(d data, m model, other *properties) {
	y := m
	y.Properties = other
	if y.Properties != nil {
		d.Set("props", *m.Properties) // want "dereference of possibly-nil `m.Properties` may panic - add a nil check or use pointer.From"
	}
}

// Should NOT be flagged: a for condition is re-checked every iteration, so a write in the
// body is fine; a local reassigned then re-created is fine too.
func validLoopCondRechecked(d data, n *node) {
	for n != nil {
		d.Set("n", *n)
		n = (*n).next
	}
}

// Should be flagged WITHOUT a fix: the value reaches a request body through another call —
// pointer.To(pointer.From(x)) would be exactly the empty write the fix must not produce.
func wrap(w widgets.Widget) *widgets.Widget { return &w }

func invalidPayloadThroughCall(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	_ = c.PutPtr(ctx, "id", pointer.To(*h.Model)) // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
	_ = c.PutPtr(ctx, "id", wrap(*h.Model))       // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
}

// Should NOT be flagged: a nested write of a provably non-nil value (a literal, or a helper
// proven never to return nil) cannot make the pointer nil again.
func validNestedNonNilReassign(d data, flag bool) {
	pr := expandLocal()
	if flag {
		pr = &properties{}
	}
	d.Set("props", *pr)
	q := expandLocal()
	if flag {
		q = expandLocal()
	}
	d.Set("props", *q)
}

// Should NOT be flagged: the nested write copies a variable that is itself proven non-nil
// where the write happens.
func validNestedGuardedReassign(d data, parse func(string) (*properties, error), flag bool) {
	pr, err := parse("x")
	if err != nil {
		return
	}
	if flag {
		next, err := parse("y")
		if err != nil {
			return
		}
		pr = next
	}
	d.Set("props", *pr)
}

// Should NOT be flagged: a nested re-assignment from a call whose error is checked right
// there keeps the value valid.
func validNestedCompanionReassign(d data, parse func(string) (*properties, error), flag bool) {
	pr, err := parse("x")
	if err != nil {
		return
	}
	if flag {
		pr, err = parse("y")
		if err != nil {
			return
		}
	}
	d.Set("props", *pr)
}

// Should be flagged: the nested re-assignment's error is not checked.
func invalidNestedUncheckedReassign(d data, parse func(string) (*properties, error), flag bool) {
	pr, err := parse("x")
	if err != nil {
		return
	}
	if flag {
		pr, _ = parse("y")
	}
	d.Set("props", *pr) // want "dereference of possibly-nil `pr` may panic - add a nil check or use pointer.From"
}

// Should NOT be flagged: the nested re-fetch of the whole struct is followed by a fresh nil
// check on the field before the block ends.
func validNestedRefetchReguarded(d data, get func() (resp, error), flag bool) {
	existing, err := get()
	if err != nil || existing.Model == nil {
		return
	}
	if flag {
		existing, err = get()
		if err != nil || existing.Model == nil {
			return
		}
	}
	d.Set("model", *existing.Model)
}

// Should be flagged: the nested re-fetch is not re-checked.
func invalidNestedRefetch(d data, get func() (resp, error), flag bool) {
	existing, err := get()
	if err != nil || existing.Model == nil {
		return
	}
	if flag {
		existing, err = get()
		if err != nil {
			return
		}
	}
	d.Set("model", *existing.Model) // want "dereference of possibly-nil `existing.Model` may panic - add a nil check or use pointer.From"
}

// Should be flagged: the err checked after the call was overwritten in between, inside a
// nested block, so it no longer speaks for the first result.
func invalidCompanionOverwrittenNested(d data, parse func(string) (*properties, error), other func() error, flag bool) {
	pr, err := parse("x")
	if flag {
		err = other()
	}
	if err != nil {
		return
	}
	d.Set("props", *pr) // want "dereference of possibly-nil `pr` may panic - add a nil check or use pointer.From"
}
