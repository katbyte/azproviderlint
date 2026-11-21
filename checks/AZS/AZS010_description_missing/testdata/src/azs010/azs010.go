package azs010

import (
	apischema "github.com/hashicorp/go-azure-sdk/resource-manager/apimanagement/2022-08-01/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// the provider reaches the plugin sdk through aliases like this one
type Schema = schema.Schema

func read() error { return nil }

func flagged() *schema.Resource {
	return &schema.Resource{ // want `resource has no description`
		Read: read,

		Schema: map[string]*schema.Schema{
			"name": { // want `property "name" has no description`
				Type:     schema.TypeString,
				Optional: true,
			},

			"sku": { // want `block "sku" has no description`
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tier": { // want `property "tier" has no description`
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"rule": { // want `block "rule" has no description`
				Type:     schema.TypeList,
				Optional: true,
				Elem:     ruleSchema(),
			},

			"zones": { // want `property "zones" has no description`
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"aliased": &Schema{ // want `property "aliased" has no description`
				Type: schema.TypeString,
			},

			"location": location(),
		},
	}
}

// a block's resource needs no description, its properties do
func ruleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": { // want `property "action" has no description`
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func location() *schema.Schema {
	return &schema.Schema{ // want `property has no description`
		Type:     schema.TypeString,
		Optional: true,
	}
}

func flaggedFramework() fwschema.Schema {
	return fwschema.Schema{ // want `schema has no description`
		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{ // want `property "name" has no description`
				Optional: true,
			},
		},

		Blocks: map[string]fwschema.Block{
			"rule": fwschema.ListNestedBlock{ // want `block "rule" has no description`
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"action": fwschema.StringAttribute{ // want `property "action" has no description`
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func flaggedFunction() function.Definition {
	return function.Definition{ // want `function has no description`
		Summary: "Parses a Thing ID",

		Parameters: []function.Parameter{
			function.StringParameter{Name: "id"}, // want `parameter has no description`
		},
	}
}

type ThingResource struct{} // want `resource ThingResource has no Description\(\) method`

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing.",
		ReadContext: read,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the Thing.",
			},

			// a blank description is there, and is another check's to report
			"blank": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},

			"sku": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The SKU of the Thing.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tier": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The tier of the SKU.",
						},
					},
				},
			},

			"zones": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The zones to use.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashSchema(&schema.Schema{Type: schema.TypeString}),
			},
		},
	}
}

func passingFramework() fwschema.Schema {
	return fwschema.Schema{
		MarkdownDescription: "Manages a Thing.",

		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the Thing.",
			},
		},

		Blocks: map[string]fwschema.Block{
			"rule": fwschema.ListNestedBlock{
				Description: "The rules of the Thing.",
			},
		},
	}
}

func passingFunction() function.Definition {
	return function.Definition{
		Description: "Parses a Thing ID.",

		Parameters: []function.Parameter{
			function.StringParameter{Name: "id", Description: "The ID to parse."},
		},
	}
}

type WidgetResource struct{}

func (WidgetResource) ResourceType() string { return "azurerm_widget" }

func (WidgetResource) Description() string { return "Manages a Widget." }

// an Azure API model is not a schema, whatever its package is called
func model() apischema.SchemaContractProperties {
	return apischema.SchemaContractProperties{}
}

type notAResource struct{}

func (notAResource) Name() string { return "something else" }
