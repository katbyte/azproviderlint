package azs021

import "azs021/pluginsdk"

// a schema made up inside a test is never shipped
func testSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"certificate": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"thumbprint": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
				},
			},
		},
	}
}
