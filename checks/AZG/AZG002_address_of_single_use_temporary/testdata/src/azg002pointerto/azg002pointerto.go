package azg002pointerto

import (
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
)

type Enum string

type props struct {
	Name  *string
	Count *int
}

// Should be flagged with pointer.To advice.
func invalidReturn() *string {
	name := "hello" // want `"name" is only used as an address by the following statement and should be inlined with pointer\.To`
	return &name
}

// Should be flagged: a conversion initializer fixes to pointer.To(Enum(s)), which AZG003
// then rewrites to pointer.ToEnum.
func invalidConversion(s string) *Enum {
	e := Enum(s) // want `"e" is only used as an address by the following statement and should be inlined with pointer\.To`
	return &e
}

// Should NOT be flagged: existing pointer.To calls are this mode's target form.
func validExisting(p *props) {
	p.Count = pointer.To(1)
}

// Should be flagged, with the aliasing hint: the temporary copies a dereferenced pointer, so
// the human may prefer returning the original pointer over the copy. No fix is offered.
func invalidDerefCopy(p props) *string {
	out := *p.Name // want `"out" is only used as an address by the following statement and should be inlined with pointer\.To - or, if sharing the original pointer is acceptable, use p\.Name directly`
	return &out
}
