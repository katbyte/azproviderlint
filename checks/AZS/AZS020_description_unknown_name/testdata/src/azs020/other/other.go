// Package other is another service of the provider, which the package under test does not import.
package other

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type GadgetResource struct{}

func (GadgetResource) ResourceType() string { return "example_gadget" }

type GadgetModel struct {
	VersionlessId string `tfschema:"versionless_id"`
}

func resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"example_widget": {
			Schema: map[string]*schema.Schema{
				"dapr": {
					Type: schema.TypeList,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{"app_id": {Type: schema.TypeString}}},
				},
			},
		},
	}
}
