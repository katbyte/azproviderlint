// Package azs015properties holds descriptions with a line break in each place one can be, for the line-break setting given as "properties".
package azs015properties

import (
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing.\nIt has parts.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name\nof the Thing.", // want `description has a line break in it - use a space`
			},

			"rule": {
				Type:        schema.TypeList,
				Description: "The rules\nof the Thing.", // want `description has a line break in it - use a space`
			},

			"zones": {
				Type:        schema.TypeSet,
				Description: "The zones\nof the Thing.", // want `description has a line break in it - use a space`
			},
		},
	}
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Description: "Manages a Thing.\nIt has parts.",
		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{
				Description: "The name\nof the Thing.", // want `description has a line break in it - use a space`
			},
		},
		Blocks: map[string]fwschema.Block{
			"rule": fwschema.ListNestedBlock{
				Description: "The rules\nof the Thing.", // want `description has a line break in it - use a space`
			},

			"zones": fwschema.SetNestedBlock{
				Description: "The zones\nof the Thing.", // want `description has a line break in it - use a space`
			},
		},
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "Manages a Thing.\nIt has parts."
}
