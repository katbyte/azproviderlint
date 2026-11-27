// Package commonschema stands in for go-azure-helpers' shared schema, which the provider uses and cannot change.
package commonschema

import "azs028/schema"

func ZonesMultipleOptional() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func Identity() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
}
