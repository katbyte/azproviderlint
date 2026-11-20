package azs012

import (
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a API Management Service.", // want `"a API" should be "an API"`

		Schema: map[string]*schema.Schema{
			"count": {
				Type:        schema.TypeString,
				Description: "A optional count, as a integer.", // want `"A optional" should be "An optional"` `"a integer" should be "an integer"`
			},

			"certificate": {
				Type:        schema.TypeString,
				Description: "The ID of an Certificate used by an user, or \"a HTTPS\" endpoint.", // want `"an Certificate" should be "a Certificate"` `"an user" should be "a user"` `"a HTTPS" should be "an HTTPS"`
			},

			"raw": {
				Type:        schema.TypeString,
				Description: `Takes a hour, or an [URL](https://example.com).`, // want `"a hour" should be "an hour"` `"an URL" should be "a URL"`
			},
		},
	}
}

func flaggedFramework() fwschema.Attribute {
	return fwschema.StringAttribute{
		MarkdownDescription: "An unique name.", // want `"An unique" should be "A unique"`
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "Manages a Automation Thing." // want `"a Automation" should be "an Automation"`
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an API Management Service.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "A user-assigned name for an SSL certificate, a URL, an hour, a one-time code or an `identity` block.",
			},

			"record": {
				Type:        schema.TypeString,
				Description: "Either a type A or an AAAA record, on a SQL or an SQL server.",
			},
		},
	}
}
