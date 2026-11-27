// Package migration holds what old state looked like, whose names cannot change.
package migration

import "azs028/pluginsdk"

func schemaV0() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"zone": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
	}
}
