package azs011

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// a test's own schema is not documentation
func inATest() *schema.Schema {
	return &schema.Schema{Description: ""}
}
