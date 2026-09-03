package azg009

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
