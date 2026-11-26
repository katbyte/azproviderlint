package azs026

import (
	fwschema "azs026/framework/schema"
	"azs026/pluginsdk"
)

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"zone_redundant": { // want `bool property "zone_redundant" should end in _enabled`
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"compression_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			// a block that is switched on by one bool
			"enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"rules": { // want `block "rules" has a plural name - a block is one thing each time it is written, so name it in the singular`
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"zone": { // want `list "zone" has a singular name - it holds many values, so name it in the plural`
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},
					},
				},
			},

			"rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{},
				},
			},

			"allowed_ip_ranges": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			// what is only read back is a list of them
			"endpoints": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{},
				},
			},

			// words that end in s and are not plurals, and plurals that do not end in s
			"dns": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{},
				},
			},

			"ip_address": { // want `list "ip_address" has a singular name`
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"user_data": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			// a name on its way out is not worth changing
			"secure": {
				Type:       pluginsdk.TypeBool,
				Optional:   true,
				Deprecated: "use `secure_enabled`",
			},

			// what a list holds cannot be read when it comes from a function
			"setting": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem:     values(),
			},
		},
	}
}

func values() *pluginsdk.Schema { return nil }

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"zone_redundant": fwschema.BoolAttribute{ // want `bool property "zone_redundant" should end in _enabled`
				Optional: true,
			},

			"compression_enabled": fwschema.BoolAttribute{
				Optional: true,
			},
		},
	}
}
