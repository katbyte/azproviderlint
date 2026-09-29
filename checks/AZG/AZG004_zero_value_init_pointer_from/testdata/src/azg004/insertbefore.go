package azg004

import (
	"fmt"
	"strings"

	// response classifies HTTP responses
	"github.com/hashicorp/go-azure-helpers/lang/response"
)

type beforeModel struct {
	Name   *string
	Status int
}

// Should be flagged: every external import sorts after the pointer package, so the new import
// goes first in the external group — above the doc comment of the import it precedes, not
// between that comment and its path.
func invalidInsertBefore(m *beforeModel) {
	name := "" // want `pointer\.From`
	if m.Name != nil {
		name = *m.Name
	}
	_ = fmt.Sprintf("%s %v", strings.TrimSpace(name), response.WasNotFound(m.Status))
}
