// Package migration holds what old state looked like, which no docs are made from.
package migration

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type ThingV0ToV1 struct{}

func (ThingV0ToV1) Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
