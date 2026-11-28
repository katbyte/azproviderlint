// Package commonschema stands in for go-azure-helpers' shared schema, which the provider uses and cannot change.
package commonschema

import (
	"azs025/schema"
	"azs025/validation"
)

func Capacity() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
	}
}

func Port() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeInt,
	}
}

func Instances() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Required:     true,
		ValidateFunc: validation.IntBetween(1, 10),
	}
}

// a string is not this rule's
func Name() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
}
