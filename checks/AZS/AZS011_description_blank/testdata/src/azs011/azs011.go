package azs011

import (
	apischema "github.com/hashicorp/go-azure-sdk/resource-manager/apimanagement/2022-08-01/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// the provider reaches the plugin sdk through aliases like this one
type Schema = schema.Schema

const empty = ""

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "", // want `description is blank - write one, or remove it`

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "", // want `description is blank - write one, or remove it`
			},

			"spaces": {
				Type:        schema.TypeString,
				Description: "  ", // want `description is blank - write one, or remove it`
			},

			"raw": {
				Type:        schema.TypeString,
				Description: ``, // want `description is blank - write one, or remove it`
			},

			"constant": {
				Type:        schema.TypeString,
				Description: empty, // want `description is blank - write one, or remove it`
			},

			"joined": {
				Type:        schema.TypeString,
				Description: empty + " " + (empty), // want `description is blank - write one, or remove it`
			},

			"aliased": &Schema{
				Type:        schema.TypeString,
				Description: "", // want `description is blank - write one, or remove it`
			},
		},
	}
}

func flaggedFramework() fwschema.Attribute {
	return fwschema.StringAttribute{
		Description:         "", // want `description is blank - write one, or remove it`
		MarkdownDescription: "", // want `description is blank - write one, or remove it`
	}
}

func flaggedFunction() function.Definition {
	return function.Definition{
		Summary:             "parse_thing_id",
		Description:         "", // want `description is blank - write one, or remove it`
		MarkdownDescription: "", // want `description is blank - write one, or remove it`

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "id",
				Description: "", // want `description is blank - write one, or remove it`
			},
		},
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "" // want `description is blank - write one, or remove it`
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Thing.",
			},

			// no description at all is another matter
			"undescribed": {
				Type: schema.TypeString,
			},
		},
	}
}

// an Azure API model's description is the user's own text, and may well be empty
func model() apischema.SchemaContractProperties {
	return apischema.SchemaContractProperties{Description: ""}
}

type notAResource struct{}

func (notAResource) Description() string { return "" }

// a block is described like any other property, and what is inside it too
func flaggedBlocks() (*schema.Schema, fwschema.Block) {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "", // want `description is blank - write one, or remove it`
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tier": {
					Type:        schema.TypeString,
					Description: " ", // want `description is blank - write one, or remove it`
				},
			},
		},
	}, fwschema.ListNestedBlock{
		MarkdownDescription: "", // want `description is blank - write one, or remove it`
	}
}
