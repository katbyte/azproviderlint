// Package options is checked with display_name and then name going first, timeouts last, and only a name ending _parent allowed among the first.
package options

import "azs027/pluginsdk"

func resource() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"thing_parent": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"workspace_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"display_name": { // want `property "display_name" should come before "name" - display_name, name go first`
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"timeouts": { // want `property "timeouts" should come after "tags" - it goes last`
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		// no longer one of the first or the last
		"tags": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"location": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},
	}
}
