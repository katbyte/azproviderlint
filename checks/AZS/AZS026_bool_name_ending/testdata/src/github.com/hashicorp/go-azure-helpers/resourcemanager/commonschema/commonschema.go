// Package commonschema stands in for go-azure-helpers' shared schema, which the provider uses and cannot change.
package commonschema

import "azs026/schema"

func ZoneRedundant() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
	}
}

func Zone() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}
}
