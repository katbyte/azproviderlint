package azg004

import (
	"strings"

	_ "github.com/hashicorp/go-azure-helpers/lang/pointer"
)

type blankConf struct {
	Description *string
}

// Should be flagged, but WITHOUT a fix: the pointer package is imported for side effects
// only, so there is no name to reference it by.
func invalidBlankPointerImport(c *blankConf) {
	description := "" // want `pointer\.From`
	if c.Description != nil {
		description = *c.Description
	}
	_ = strings.TrimSpace(description)
}
