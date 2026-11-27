// Package options is checked with bools allowed to end in _enabled or _required only.
package options

import "azs026/pluginsdk"

func resource() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"mtls_required": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"terms_of_service_accepted": { // want `bool property "terms_of_service_accepted" should end in _enabled or _required`
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},
	}
}
