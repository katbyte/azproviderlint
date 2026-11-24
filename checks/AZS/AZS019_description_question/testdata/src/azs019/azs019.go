package azs019

import (
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flagged() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// read-only, so it is what it is and not what it should be
			"https_required": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Should HTTPS be required?", // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
			},

			"audit_log_enabled": {
				Type:        schema.TypeBool,
				Description: "Should the log auditing be enabled?", // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
			},

			"web_sockets_enabled": {
				Type:        schema.TypeBool,
				Description: "Should Web Sockets be enabled. Defaults to `false`.", // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
			},

			// only "Should ... be ..." has a fix
			"https_only": {
				Type:        schema.TypeBool,
				Description: "Is the Function App only accessible via HTTPS?", // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
			},

			"use_32_bit_worker": {
				Type:        schema.TypeBool,
				Description: "Does the Linux Web App use a 32-bit worker.", // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
			},

			"gateway_disabled": {
				Type:        schema.TypeBool,
				Description: `Disable the gateway in the main region? This is only supported when "additional_location" is set.`, // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
			},
		},
	}
}

func flaggedFramework() fwschema.Attribute {
	return fwschema.StringAttribute{
		MarkdownDescription: "(Boolean) Should the default recipients be notified?", // want `description asks a question - say what it is, such as "Whether the Thing is enabled\."`
	}
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Thing.",

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the Thing is enabled.",
			},

			"count": {
				Type:        schema.TypeString,
				Description: "Must be between `1` and `5`.",
			},

			"mode": {
				Type:        schema.TypeString,
				Description: "Is required when `sku` is set.",
			},

			"query": {
				Type:        schema.TypeString,
				Description: "The query to run, such as `enabled?`.",
			},
		},
	}
}
