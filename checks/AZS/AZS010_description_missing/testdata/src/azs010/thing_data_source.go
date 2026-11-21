package azs010

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSource() *schema.Resource {
	return &schema.Resource{ // want `data source has no description`
		Read: read,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Thing.",
			},
		},
	}
}

type ThingDataSource struct{} // want `data source ThingDataSource has no Description\(\) method`

func (ThingDataSource) ResourceType() string { return "azurerm_thing" }
