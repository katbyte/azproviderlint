package azg011

type sku struct {
	Name *string
}

type properties struct {
	Count *int
	Sku   *sku
}

func (properties) label() string     { return "" }
func (*properties) ptrLabel() string { return "" }

type model struct {
	Properties *properties
}

func use(interface{}) {}

// Should be flagged: selecting Count implicitly dereferences the possibly-nil Properties.
func invalidImplicitChain(m model) {
	use(m.Properties.Count) // want "selecting `Count` implicitly dereferences possibly-nil `m.Properties` and may panic - add a nil check"
}

// Should be flagged twice: both Properties and Sku links are unguarded.
func invalidDeepChain(m model) {
	use(m.Properties.Sku.Name) // want "selecting `Name` implicitly dereferences possibly-nil `m.Properties.Sku` and may panic - add a nil check" "selecting `Sku` implicitly dereferences possibly-nil `m.Properties` and may panic - add a nil check"
}

// Should be flagged once: the guard covers Properties but not the deeper Sku link.
func invalidPartialGuard(m model) {
	if m.Properties != nil {
		use(m.Properties.Sku.Name) // want "selecting `Name` implicitly dereferences possibly-nil `m.Properties.Sku` and may panic - add a nil check"
	}
}

// Should be flagged: the guard condition itself dereferences the unchecked prefix.
func invalidGuardInCondition(m model) {
	if m.Properties.Count != nil { // want "selecting `Count` implicitly dereferences possibly-nil `m.Properties` and may panic - add a nil check"
		use(1)
	}
}

// Should be flagged: a value-receiver method call dereferences the pointer.
func invalidValueReceiver(m model) {
	use(m.Properties.label()) // want "selecting `label` implicitly dereferences possibly-nil `m.Properties` and may panic - add a nil check"
}

// Should NOT be flagged: a pointer-receiver method is called on the pointer itself.
func validPtrReceiver(m model) {
	use(m.Properties.ptrLabel())
}

// Should NOT be flagged: enclosing if proves the link non-nil.
func validGuardedImplicit(m model) {
	if m.Properties != nil {
		use(m.Properties.Count)
	}
}

// Should NOT be flagged: the && left operand proves the link non-nil for the right.
func validShortCircuit(m model) bool {
	return m.Properties != nil && m.Properties.Count != nil
}

// Should NOT be flagged: an early return guards everything below it.
func validEarlyReturn(m model) {
	if m.Properties == nil {
		return
	}
	use(m.Properties.Count)
}

// Should NOT be flagged: a bare pointer parameter's nil contract belongs to its callers.
func validBareParam(props *properties) {
	use(props.Count)
}

// Should be flagged: deeper links through a parameter are always in scope.
func invalidParamField(props *properties) {
	use(props.Sku.Name) // want "selecting `Name` implicitly dereferences possibly-nil `props.Sku` and may panic - add a nil check"
}

// Should be flagged: an assignment-target dereference panics and pointer.From cannot help.
func invalidWriteTarget(props properties) {
	*props.Count = 1 // want "dereference of possibly-nil `props.Count` may panic - add a nil check"
}

// Should be flagged: &*x panics on nil x just like *x.
func invalidAddressOfDeref(props properties) {
	use(&*props.Count) // want "dereference of possibly-nil `props.Count` may panic - add a nil check"
}

// Should be flagged: an inc/dec statement needs the pointer, not a copy.
func invalidIncDec(props properties) {
	(*props.Count)++ // want "dereference of possibly-nil `props.Count` may panic - add a nil check"
}

// Should NOT be flagged: the guard covers the write target.
func validGuardedWrite(props properties) {
	if props.Count != nil {
		*props.Count = 1
	}
}

// Should NOT be flagged here: a value-context dereference is AZG008's, fixable by pointer.From.
func valueDerefIsAZG008(props properties) int {
	return *props.Count
}

// Should be flagged: reassignment inside the guarded body invalidates the enclosing guard.
func invalidReassignedInsideGuard(m model, next func() *properties) {
	if m.Properties != nil {
		m.Properties = next()
		use(m.Properties.Count) // want "selecting `Count` implicitly dereferences possibly-nil `m.Properties` and may panic - add a nil check"
	}
}

func lookup() (*properties, bool) { return nil, false }

func fetch() (*properties, error) { return nil, nil }

// Should NOT be flagged: the ok companion is checked on the left of && in the same condition.
func validOkInCondition() bool {
	if props, ok := lookup(); ok && props.Count != nil {
		return true
	}
	return false
}

// Should NOT be flagged: `err != nil ||` proves the call succeeded for the right operand.
func validErrInCondition() bool {
	props, err := fetch()
	return err != nil || props.Count == nil
}

// Should be flagged: `ok ||` runs the right operand when ok is false.
func invalidOkOrInCondition() bool {
	props, ok := lookup()
	return ok || props.Count == nil // want "selecting `Count` implicitly dereferences possibly-nil `props` and may panic - add a nil check"
}

// Should NOT be flagged: a closure's pointer parameter is trusted like any other parameter.
func validClosureParam() {
	f := func(props *properties) { use(props.Count) }
	f(nil)
}

// Should be flagged: deeper links through a closure parameter stay in scope.
func invalidClosureParamField() {
	f := func(props *properties) { use(props.Sku.Name) } // want "selecting `Name` implicitly dereferences possibly-nil `props.Sku` and may panic - add a nil check"
	f(nil)
}

type holder struct {
	Inner *properties
	Tags  *map[string]string
	Array *[3]int
}

// Should be flagged: writing a field through a dereference needs the pointer itself.
func invalidFieldWrite(h holder) {
	(*h.Inner).Count = nil // want "dereference of possibly-nil `h.Inner` may panic - add a nil check"
}

// Should be flagged: writing a map entry through a dereference panics on a nil pointer.
func invalidMapIndexWrite(h holder) {
	(*h.Tags)["k"] = "v" // want "dereference of possibly-nil `h.Tags` may panic - add a nil check"
}

// Should be flagged: writing an array element through a dereference needs the array itself.
func invalidArrayIndexWrite(h holder) {
	(*h.Array)[0] = 1 // want "dereference of possibly-nil `h.Array` may panic - add a nil check"
}

// Should be flagged: slicing an array needs the array itself, not a copy.
func invalidArraySlice(h holder) {
	use((*h.Array)[:]) // want "dereference of possibly-nil `h.Array` may panic - add a nil check"
}

// Should be flagged: the address of a field reached through a dereference.
func invalidAddressOfField(h holder) {
	use(&(*h.Inner).Count) // want "dereference of possibly-nil `h.Inner` may panic - add a nil check"
}

// Should be flagged: a pointer-receiver method called on a dereference needs the pointer.
func invalidPtrReceiverOnDeref(h holder) {
	use((*h.Inner).ptrLabel()) // want "dereference of possibly-nil `h.Inner` may panic - add a nil check"
}

// Should NOT be flagged: each of those is covered by a guard on the same chain.
func validGuardedPointerContexts(h holder) {
	if h.Inner != nil {
		(*h.Inner).Count = nil
		use(&(*h.Inner).Count)
	}
	if h.Tags == nil || h.Array == nil {
		return
	}
	(*h.Tags)["k"] = "v"
	(*h.Array)[0] = 1
}

// Should NOT be flagged: a method expression names a type, so nothing is dereferenced.
func validMethodExpression() {
	use((*properties).ptrLabel)
	use(properties.label)
}
