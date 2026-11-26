package azs027

import "azs027/pluginsdk"

// a data source reads these back, so its order is its own
func dataSource() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"tags": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}
