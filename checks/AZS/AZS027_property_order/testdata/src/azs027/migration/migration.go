// Package migration holds what old state looked like, in whatever order it was written.
package migration

import "azs027/pluginsdk"

func schemaV0() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"tags": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}
