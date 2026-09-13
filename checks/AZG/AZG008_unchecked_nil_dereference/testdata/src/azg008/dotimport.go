package azg008

import (
	"context"
	. "net/http"

	"github.com/example/sdk/widgets"
)

// Should be flagged WITHOUT a fix: the write method is a bare identifier from a dot-imported
// net/http.
func invalidDotImportedMethod(c widgets.WidgetsClient, ctx context.Context, h widgetHolder) {
	_ = c.Do(ctx, MethodPost, "id", *h.Model) // want "may panic - add a nil check \\(no fix: the value is sent as a write request body"
}
