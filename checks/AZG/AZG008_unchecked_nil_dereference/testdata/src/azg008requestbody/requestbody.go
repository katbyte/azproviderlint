package azg008requestbody

import (
	"context"

	"github.com/example/sdk/widgets"
)

type widgetHolder struct{ Model *widgets.Widget }

type data struct{}

func (data) Set(string, interface{}) {}

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

// Reported WITH a fix even here: a GET sends no body, and a copy that never reaches a write is
// an ordinary dereference.
func invalidNotPayload(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	_ = c.Get(ctx, "id", *h.Model) // want "may panic - add a nil check or use pointer.From"
}

func invalidCopyNotSent(d data, h widgetHolder) {
	env := widgets.Envelope{Widget: *h.Model} // want "may panic - add a nil check or use pointer.From"
	d.Set("env", env)
}
