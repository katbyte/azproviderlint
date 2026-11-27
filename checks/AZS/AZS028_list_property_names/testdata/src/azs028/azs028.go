package azs028

import (
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"

	"azs028/pluginsdk"
)

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
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

			// a product's name, a word with no other form, and a word that is not a noun
			"application_insights": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{},
				},
			},

			"location_data": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{},
				},
			},

			"string_contains": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{},
				},
			},

			// the plural is not always the last word, and does not always end in s
			"days_of_week": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"tables_to_include": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"local_users_not_allowed": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"severities_allowed": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"values_base64": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"outbound_ip_address_list": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"provisionable_cpus": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"day_of_week": { // want `list "day_of_week" has a singular name`
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
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

			// what a list holds is told by the type a function hands back
			"setting": { // want `list "setting" has a singular name`
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem:     values(),
			},

			"members": { // want `block "members" has a plural name`
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem:     member(),
			},

			// a list whose schema comes from a function, in the provider or outside it, is named where it is used
			"zone":          commonschema.ZonesMultipleOptional(), // want `list "zone" has a singular name`
			"identities":    commonschema.Identity(),              // want `block "identities" has a plural name`
			"peer_address":  addresses(),                          // want `list "peer_address" has a singular name`
			"zones_from_us": commonschema.ZonesMultipleOptional(),
		},
	}
}

func values() *pluginsdk.Schema { return nil }

func member() *pluginsdk.Resource { return nil }

func addresses() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Schema{
			Type: pluginsdk.TypeString,
		},
	}
}
