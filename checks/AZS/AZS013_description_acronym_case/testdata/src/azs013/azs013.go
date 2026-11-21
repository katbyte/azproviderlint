package azs013

import (
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an Api Management Service.", // want `"Api" should be "API"`

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "The id of the subnet, and the ids of its vms.", // want `"id" should be "ID"` `"ids" should be "IDs"` `"vms" should be "VMs"`
			},

			"endpoint": {
				Type:        schema.TypeString,
				Description: "The url (http/https) or the ip's Dns name, as Json.", // want `"url" should be "URL"` `"http" should be "HTTP"` `"https" should be "HTTPS"` `"ip" should be "IP"` `"Dns" should be "DNS"` `"Json" should be "JSON"`
			},

			"raw": {
				Type:        schema.TypeString,
				Description: `The [Client Id](https://example.com/id) for OAUTH on the vnet.`, // want `"Id" should be "ID"` `"OAUTH" should be "OAuth"` `"vnet" should be "VNet"`
			},
		},
	}
}

func flaggedFramework() fwschema.Attribute {
	return fwschema.StringAttribute{
		MarkdownDescription: "The Sku of the Thing.", // want `"Sku" should be "SKU"`
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "Manages a Sql Thing." // want `"Sql" should be "SQL"`
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an API Management Service.",

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "The ID of the subnet, as returned by `azurerm_subnet.example.id`, with its IDs, VMs and IPv4 addresses.",
			},

			"endpoint": {
				Type:        schema.TypeString,
				Description: "The URL, such as https://example.com/api/id, in ssh-rsa format, made with \"ssh-keygen -t rsa\" and sent as api-version or example.id.",
			},
		},
	}
}

// a block is described like any other property, and what is inside it too
func flaggedBlocks() (*schema.Schema, fwschema.Block) {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "A Sku block.", // want `"Sku" should be "SKU"`
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tier": {
					Type:        schema.TypeString,
					Description: "The id of the tier.", // want `"id" should be "ID"`
				},
			},
		},
	}, fwschema.ListNestedBlock{
		MarkdownDescription: "A rule block for an ip range.", // want `"ip" should be "IP"`
	}
}
