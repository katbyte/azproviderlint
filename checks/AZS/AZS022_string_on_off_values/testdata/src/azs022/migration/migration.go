// Package migration holds what old state looked like, which cannot be given another type.
package migration

import (
	"azs022/pluginsdk"
	"azs022/validation"
)

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"public_access": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
		},
	}
}
