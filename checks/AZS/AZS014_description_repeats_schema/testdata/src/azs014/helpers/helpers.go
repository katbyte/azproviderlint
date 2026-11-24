// Package helpers holds schema shared between resources, as a provider's service packages do.
package helpers

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// NameSchema builds the name property every child resource has, around the description it is given.
func NameSchema(description string) *schema.Schema {
	return &schema.Schema{
		Description: description,
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}
}
