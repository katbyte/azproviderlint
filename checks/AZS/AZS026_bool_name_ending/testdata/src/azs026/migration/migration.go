// Package migration holds what old state looked like, whose names cannot change.
package migration

import "azs026/pluginsdk"

func schemaV0() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"zone_redundant": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},
	}
}
