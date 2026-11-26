// Package options is checked with blocks switched off, bools allowed to end in _required or _enabled, "settings" not a plural, "sheep" one, and "dark" allowed after the plural.
package options

import "azs026/pluginsdk"

func resource() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"mtls_required": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"zone_redundant": { // want `bool property "zone_redundant" should end in _required or _enabled`
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},

		"rules": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{},
			},
		},

		"settings": { // want `list "settings" has a singular name`
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"sheep": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"preview_images_dark": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
	}
}
