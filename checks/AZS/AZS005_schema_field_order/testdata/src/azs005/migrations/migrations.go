// Package migrations holds what old state looked like, which is left alone when state-migrations is false.
package migrations

import "azs005/pluginsdk"

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Optional: true,
			Type:     pluginsdk.TypeString,
		},
	}
}
