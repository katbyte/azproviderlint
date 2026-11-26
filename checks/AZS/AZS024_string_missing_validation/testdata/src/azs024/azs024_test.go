package azs024

import "azs024/pluginsdk"

// a schema made up inside a test is never shipped
func testSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}
