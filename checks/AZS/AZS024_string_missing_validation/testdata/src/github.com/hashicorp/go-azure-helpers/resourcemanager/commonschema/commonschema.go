// Package commonschema stands in for go-azure-helpers' shared schema, which the provider uses and cannot change.
package commonschema

import (
	"azs024/schema"
	"azs024/validation"
)

func ZoneSingleOptionalComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	}
}

// one handed on from another is followed
func ZoneSingleOptional() *schema.Schema {
	return ZoneSingleOptionalComputed()
}

func ZonesMultipleOptional() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func Zone() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeString,
	}
}

func Location() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 90),
	}
}

func Name() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
	}
}

func ZoneComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}
