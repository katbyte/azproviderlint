package azs017

import (
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flagged() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Specifies the name of the Webhook.", // want `description opens with "Specifies the" - open with "The"`
			},

			"scopes": {
				Type:        schema.TypeString,
				Description: "Specifies a list of scopes to be requested as part of Facebook Login authentication.", // want `description opens with "Specifies a" - open with "A"`
			},

			"http2_enabled": {
				Type:        schema.TypeString,
				Description: "Specifies if the HTTP2 protocol should be enabled.", // want `description opens with "Specifies if" - open with "Whether"`
			},

			"tracing_enabled": {
				Type:        schema.TypeString,
				Description: `Indicates whether tracing is enabled.`, // want `description opens with "Indicates whether" - open with "Whether"`
			},

			"hybrid_worker_id": {
				Type:        schema.TypeString,
				Description: "Specify the ID of this Hybrid Worker in UUID notation.", // want `description opens with "Specify the" - open with "The"`
			},

			// what follows the verb cannot open the description as it stands, so there is no fix
			"interval": {
				Type:        schema.TypeString,
				Description: "Specifies interval in minutes to reevaluate the percentage.", // want `description opens with "Specifies" - say what the property is instead`
			},
		},
	}
}

func flaggedFramework() fwschema.Attribute {
	return fwschema.StringAttribute{
		MarkdownDescription: "Defines how the storage adapters between nodes are connected.", // want `description opens with "Defines how" - open with "How"`
	}
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Webhook.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Webhook. Specifies the Webhook within the Automation Account.",
			},

			"minimum_tls_version": {
				Type:        schema.TypeString,
				Description: "Configures the minimum version of TLS required for requests to the site.",
			},

			"enabled": {
				Type:        schema.TypeString,
				Description: "Whether the Webhook is enabled.",
			},
		},
	}
}

// a block is described like any other property
func flaggedBlocks() (*schema.Schema, fwschema.Block) {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Specifies an identity block.", // want `description opens with "Specifies an" - open with "An"`
	}, fwschema.ListNestedBlock{
		MarkdownDescription: "Specifies which rules apply.", // want `description opens with "Specifies which" - open with "Which"`
	}
}
