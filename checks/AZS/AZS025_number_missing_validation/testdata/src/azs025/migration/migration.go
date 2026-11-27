// Package migration holds what old state looked like, which nobody sets.
package migration

import "azs025/pluginsdk"

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"count": {
			Type:     pluginsdk.TypeInt,
			Required: true,
		},
	}
}
