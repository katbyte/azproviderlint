package azr010

import "github.com/example/helpers"

type sku struct{ Name *string }

type properties struct {
	Ids   *[]string
	Sku   *sku
	Count *int
}

type data struct{}

func (data) Set(string, interface{}) {}

// flattenIds assumes non-nil input.
func flattenIds(input *[]string) []interface{} {
	return helpers.FlattenStringSlice(input)
}

// flattenSku assumes non-nil input.
func flattenSku(input *sku) []interface{} {
	return []interface{}{map[string]interface{}{"name": *input.Name}}
}

// flattenSkuSafe returns early on nil.
func flattenSkuSafe(input *sku) []interface{} {
	if input == nil {
		return []interface{}{}
	}
	return flattenSku(input)
}

// flattenSkuSafeOr guards through an || chain after setting up its output.
func flattenSkuSafeOr(input *sku) []interface{} {
	out := make([]interface{}, 0)
	if input == nil || input.Name == nil {
		return out
	}
	return append(out, *input.Name)
}

// flattenSkuWrapped handles nil by wrapping every use in a guard.
func flattenSkuWrapped(input *sku) []interface{} {
	out := make([]interface{}, 0)
	if input != nil {
		out = append(out, *input.Name)
	}
	return out
}

// flattenSkuPassThrough hands its input to a flatten that does not handle nil.
func flattenSkuPassThrough(input *sku) []interface{} {
	return flattenSku(input)
}

// Should be flagged: the caller guards what the flatten should.
func invalidGuardedCall(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSku(props.Sku)) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged: the guard is on an alias defined in the if's init.
func invalidAliasInit(d data, props properties) {
	if v := props.Sku; v != nil {
		d.Set("sku", flattenSku(v)) // want "`flattenSku\\(v\\)` is called under a nil check on `v` - handle nil inside the flatten function instead"
	}
}

// Should be flagged: the else branch of a nil check is the guarded one.
func invalidElseBranch(d data, props properties) {
	if props.Sku == nil {
		d.Set("sku", []interface{}{})
	} else {
		d.Set("sku", flattenSku(props.Sku)) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged: the guard is one conjunct of a compound condition.
func invalidCompound(d data, props properties, enabled bool) {
	if enabled && props.Sku != nil {
		d.Set("sku", flattenSku(props.Sku)) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged with a fix: the callee already handles nil, so the check goes.
func invalidRedundantCheck(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuSafe(props.Sku)) // want "`flattenSkuSafe` already handles a nil `props.Sku` - drop the nil check"
	}
}

// Should be flagged with a fix: the || guard counts, and the fact crosses the delegating
// flattenIds into the helpers package.
func invalidRedundantCheckDelegating(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuSafeOr(props.Sku)) // want "`flattenSkuSafeOr` already handles a nil `props.Sku` - drop the nil check"
	}
	if props.Ids != nil {
		d.Set("ids", flattenIds(props.Ids)) // want "`flattenIds` already handles a nil `props.Ids` - drop the nil check"
	}
}

// Should be flagged without a fix: the compound condition cannot simply be dropped.
func invalidRedundantCompound(d data, props properties, enabled bool) {
	if enabled && props.Sku != nil {
		d.Set("sku", flattenSkuSafe(props.Sku)) // want "`flattenSkuSafe` already handles a nil `props.Sku` - drop the nil check"
	}
}

// Should NOT be flagged: the branch dereferences the chain too, so the check is needed.
func validOtherUse(d data, props properties) {
	if props.Sku != nil {
		d.Set("name", *props.Sku.Name)
		d.Set("sku", flattenSku(props.Sku))
	}
}

// Should NOT be flagged: the check is on a different chain.
func validGuardElsewhere(d data, props properties) {
	if props.Count != nil {
		d.Set("sku", flattenSku(props.Sku))
	}
}

// Should NOT be flagged: no guard at all is what the rule wants.
func validUnguarded(d data, props properties) {
	d.Set("sku", flattenSkuSafe(props.Sku))
}

// Should NOT be flagged: the argument is dereferenced, which is a different check's concern.
func validDerefArg(d data, props properties) {
	if props.Ids != nil {
		d.Set("ids", flattenValues(*props.Ids))
	}
}

func flattenValues(input []string) []interface{} { return nil }

// Should be flagged with a fix: a wrapping guard handles nil as well as an early return.
func invalidRedundantWrapped(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuWrapped(props.Sku)) // want "`flattenSkuWrapped` already handles a nil `props.Sku` - drop the nil check"
	}
}

// Should be flagged: passing the input to a flatten that does not handle nil is not handling it.
func invalidPassThrough(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuPassThrough(props.Sku)) // want "`flattenSkuPassThrough\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

type holder struct{ Sku *sku }

// flattenSkuParen compares the parenthesised parameter with nil, which still counts.
func flattenSkuParen(input *sku) []interface{} {
	if (input) == nil {
		return []interface{}{}
	}
	return flattenSku(input)
}

// flattenSkuStored only stores its input, never dereferencing it.
func flattenSkuStored(input *sku) holder {
	h := holder{}
	h.Sku = input
	return h
}

// flattenSkuAlias guards an alias, not the parameter: the alias is not followed.
func flattenSkuAlias(input *sku) []interface{} {
	v := input
	if v == nil {
		return []interface{}{}
	}
	return []interface{}{*v.Name}
}

func getSku() *sku { return nil }

// Should be flagged: the alias in the init is checked while the chain itself is passed.
func invalidAliasInitChain(d data, props properties) {
	if v := props.Sku; v != nil {
		d.Set("sku", flattenSku(props.Sku)) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged: nil on the left of the comparison.
func invalidNilOnLeft(d data, props properties) {
	if nil != props.Sku {
		d.Set("sku", flattenSku(props.Sku)) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged: the other conjunct is a comparison that is not a nil check.
func invalidOtherComparison(d data, props properties, n int) {
	if n > 0 && props.Sku != nil {
		d.Set("sku", flattenSku(props.Sku)) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged: a parenthesised argument is still the flatten argument itself.
func invalidParenArg(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSku((props.Sku))) // want "`flattenSku\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should NOT be flagged: comparing two chains is not a nil check.
func validCompareChains(d data, props, other properties) {
	if props.Sku != other.Sku {
		d.Set("sku", flattenSku(props.Sku))
	}
}

// Should NOT be flagged: a call result has no chain to guard.
func validCallArg(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSku(getSku()))
	}
}

// Should be flagged with a fix: a parenthesised nil comparison in the callee counts.
func invalidRedundantParen(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuParen(props.Sku)) // want "`flattenSkuParen` already handles a nil `props.Sku` - drop the nil check"
	}
}

// Should be flagged with a fix: storing the input handles nil.
func invalidRedundantStored(d data, props properties) {
	if props.Sku != nil {
		_ = flattenSkuStored(props.Sku) // want "`flattenSkuStored` already handles a nil `props.Sku` - drop the nil check"
	}
}

// Should be flagged without a fix: the callee guards an alias, which is not followed.
func invalidAliasNotFollowed(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuAlias(props.Sku)) // want "`flattenSkuAlias\\(props.Sku\\)` is called under a nil check on `props.Sku` - handle nil inside the flatten function instead"
	}
}

// Should be flagged without a fix: a single-line if cannot be unwrapped cleanly.
func invalidRedundantSingleLine(d data, props properties) {
	if props.Sku != nil { d.Set("sku", flattenSkuSafe(props.Sku)) } // want "`flattenSkuSafe` already handles a nil `props.Sku` - drop the nil check"
}

// Should be flagged with a fix: continuation lines lose one level of indentation.
func invalidRedundantMultiLine(d data, props properties) {
	if props.Sku != nil {
		d.Set(
			"sku",
			flattenSkuSafe(props.Sku), // want "`flattenSkuSafe` already handles a nil `props.Sku` - drop the nil check"
		)
	}
}

// flattenSkuBoxed places its input in a literal without dereferencing it.
func flattenSkuBoxed(input *sku) []interface{} {
	return []interface{}{input}
}

// Should be flagged with a fix: boxing the input into a literal handles nil.
func invalidRedundantBoxed(d data, props properties) {
	if props.Sku != nil {
		d.Set("sku", flattenSkuBoxed(props.Sku)) // want "`flattenSkuBoxed` already handles a nil `props.Sku` - drop the nil check"
	}
}
