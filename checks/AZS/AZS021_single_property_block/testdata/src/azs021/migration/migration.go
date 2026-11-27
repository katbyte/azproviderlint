// Package migration holds what old state looked like, which cannot be given another shape.
package migration

import "azs021/pluginsdk"

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"certificate": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"thumbprint": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
				},
			},
		},
	}
}
