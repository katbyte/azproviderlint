// Package migration holds what old state looked like, names and all.
package migration

import "azs023/pluginsdk"

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"is_enabled": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},
	}
}
