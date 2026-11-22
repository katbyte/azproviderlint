package azs016

import (
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flagged() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a function app.", // want `"function app" should be "Function App"`

		Schema: map[string]*schema.Schema{
			"storage_account_name": {
				Type:        schema.TypeString,
				Description: "The backend storage account name which will be used by this Function app, in the same resource group.", // want `"storage account" should be "Storage Account"` `"Function app" should be "Function App"` `"resource group" should be "Resource Group"`
			},

			"daily_memory_time_quota": {
				Type:        schema.TypeString,
				Description: "The memory this app may use per day. Setting this value only affects function apps in app service plans using managed identities.", // want `"function apps" should be "Function Apps"` `"app service plans" should be "App Service Plans"` `"managed identities" should be "Managed Identities"`
			},

			"virtual_network_subnet_id": {
				Type:        schema.TypeString,
				Description: `The subnet used for [regional virtual network integration](https://example.com/virtual-network/function-app) by the function app's workers.`, // want `"virtual network" should be "Virtual Network"` `"function app" should be "Function App"`
			},
		},
	}
}

func flaggedFramework() fwschema.Attribute {
	return fwschema.StringAttribute{
		MarkdownDescription: "The ID of the key vault secret.", // want `"key vault secret" should be "Key Vault Secret"`
	}
}

type ThingResource struct{}

func (ThingResource) ResourceType() string { return "azurerm_thing" }

func (ThingResource) Description() string {
	return "Manages a Thing within a log analytics Workspace." // want `"log analytics Workspace" should be "Log Analytics Workspace"`
}

func passing() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Function App.",

		Schema: map[string]*schema.Schema{
			"storage_account_name": {
				Type:        schema.TypeString,
				Description: "The backend Storage Account name used by this Function App's workers, and by other Function Apps in the Resource Group.",
			},

			"kind": {
				Type:        schema.TypeString,
				Description: "The kind of app, such as `function app` or my-function app, as shown at https://example.com/function app.",
			},

			"health_check_eviction_time": {
				Type:        schema.TypeString,
				Description: "The time in minutes a node can be unhealthy before being removed from the load balancer, after its storage accounting is done.",
			},
		},
	}
}

// a block is described like any other property, and what is inside it too
func flaggedBlocks() (*schema.Schema, fwschema.Block) {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "A storage account block.", // want `"storage account" should be "Storage Account"`
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"access_key": {
					Type:        schema.TypeString,
					Description: "The access key of the Storage account.", // want `"Storage account" should be "Storage Account"`
				},
			},
		},
	}, fwschema.ListNestedBlock{
		MarkdownDescription: "A rule block for a virtual machine.", // want `"virtual machine" should be "Virtual Machine"`
	}
}
