package azr006

import "context"

type Client struct {
	StopContext context.Context
}

type ResourceData struct{}

func ForCreate(ctx context.Context, d *ResourceData) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}

// Should be flagged: ctx assigned directly from the meta object
func badRead(d *ResourceData, meta interface{}) error {
	ctx := meta.(*Client).StopContext // want `use a timeouts-wrapped StopContext \(timeouts\.ForCreate/ForCreateUpdate/ForRead/ForUpdate/ForDelete\) so Custom Timeouts are supported, instead of assigning ctx from meta directly`
	_ = ctx
	return nil
}

// Should NOT be flagged: wrapped StopContext with cancel
func goodCreate(d *ResourceData, meta interface{}) error {
	ctx, cancel := ForCreate(meta.(*Client).StopContext, d)
	defer cancel()
	_ = ctx
	return nil
}

// Should NOT be flagged: ctx from context package
func goodBackground() {
	ctx := context.Background()
	_ = ctx
}

// Should NOT be flagged: different variable name reading from meta
func goodOtherName(meta interface{}) {
	stopCtx := meta.(*Client).StopContext
	_ = stopCtx
}

// Should be flagged: parentheses around the meta assertion do not hide it
func badParenRead(meta interface{}) {
	ctx := (meta.(*Client)).StopContext // want `use a timeouts-wrapped StopContext`
	_ = ctx
}

// Should be flagged: indexing into an asserted slice still roots at meta
func badIndexRead(meta interface{}) {
	ctx := meta.([]*Client)[0].StopContext // want `use a timeouts-wrapped StopContext`
	_ = ctx
}

// Should NOT be flagged: the chain roots at a composite literal, not an identifier
func goodLiteralRoot() {
	ctx := Client{}.StopContext
	_ = ctx
}

// Should NOT be flagged: a plain assignment, not a short declaration
func goodAssign(meta interface{}) {
	var ctx context.Context
	ctx = meta.(*Client).StopContext
	_ = ctx
}

// Should NOT be flagged: two values on the left
func goodTwoValues(meta interface{}) {
	ctx, ok := meta.(*Client).StopContext, true
	_, _ = ctx, ok
}
