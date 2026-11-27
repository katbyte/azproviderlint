// Package commonschema stands in for go-azure-helpers' shared schema, which the provider uses and cannot change.
package commonschema

import "azs027/schema"

func LocationComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}

func Tags() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
	}
}
