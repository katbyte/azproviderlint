// Package migration holds what old state looked like, which nobody writes a block for.
package migration

import "azs003/schema"

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"settings": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"foo": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}
