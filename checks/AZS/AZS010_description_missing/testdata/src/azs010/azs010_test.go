package azs010

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// a schema built in a test documents nothing
func testSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type: schema.TypeString,
		},
	}
}
