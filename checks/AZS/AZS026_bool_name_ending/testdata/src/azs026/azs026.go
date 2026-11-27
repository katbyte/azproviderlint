package azs026

import (
	fwschema "azs026/framework/schema"
	"azs026/pluginsdk"
)

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"zone_redundant": { // want `bool property "zone_redundant" should end in _enabled, _required, _accepted, _allowed, _supported, _verified, _configured or _only`
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

			// a word for its state reads better than _enabled
			"mtls_required": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"terms_of_service_accepted": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"https_only": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			// a bool the API calls disable is turned around and named for what it turns on
			"public_access_disabled": { // want `bool property "public_access_disabled" should end in _enabled`
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			// what is not a bool is not this rule's
			"zone": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			// a name on its way out is not worth changing
			"secure": {
				Type:       pluginsdk.TypeBool,
				Optional:   true,
				Deprecated: "use `secure_enabled`",
			},
		},
	}
}

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
