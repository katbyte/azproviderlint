package claims

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the Thing, please.",
			},
		},
	}
}
