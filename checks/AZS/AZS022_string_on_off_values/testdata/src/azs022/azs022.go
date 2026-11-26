package azs022

import (
	fwschema "azs022/framework/schema"
	"azs022/framework/stringvalidator"
	"azs022/pluginsdk"
	"azs022/things"
	"azs022/validation"
)

const off = "Off"

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"public_access": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false), // want `string property "public_access" only takes "Enabled" and "Disabled" - make it a bool`
			},

			// the values of an SDK enum, written out or from its helper
			"network_access": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{ // want `string property "network_access" only takes "Disabled" and "Enabled" - make it a bool`
					string(things.PublicAccessDisabled),
					string(things.PublicAccessEnabled),
				}, false),
			},

			"access": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(things.PossibleValuesForPublicAccess(), false), // want `string property "access" only takes "Disabled" and "Enabled" - make it a bool`
			},

			"logging": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInEnumSlice(things.PossibleLoggingValues(), false), // want `string property "logging" only takes "Off" and "On" - make it a bool`
			},

			// whatever the case, and whichever comes first
			"audit": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{off, "ON"}, true), // want `string property "audit" only takes "Off" and "ON" - make it a bool`
			},

			"backup": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Disable", "Enable"}, false), // want `string property "backup" only takes "Disable" and "Enable" - make it a bool`
			},

			// a third value is more than a bool can say
			"encryption": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled", "None"}, false),
			},

			"tier": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(things.PossibleValuesForTier(), false),
			},

			// one value alone, or two that do not go together
			"state": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled"}, false),
			},

			"mode": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Off"}, false),
			},

			// values that are not written out could be anything
			"replication": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(replication(), false),
			},

			"region": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(things.PossibleRegions(), false),
			},

			"zone": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", zone()}, false),
			},

			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			// a property on its way out is left alone
			"soft_delete": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Deprecated:   "use `soft_delete_enabled`",
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
			},
		},
	}
}

func replication() []string { return nil }

func zone() string { return "" }

func property() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:         pluginsdk.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice([]string{"On", "Off"}, false), // want `string property only takes "On" and "Off" - make it a bool`
	}
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"public_access": fwschema.StringAttribute{
				Optional: true,
				Validators: []any{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.OneOf("Enabled", "Disabled"), // want `string property "public_access" only takes "Enabled" and "Disabled" - make it a bool`
				},
			},

			"logging": fwschema.StringAttribute{
				Optional:   true,
				Validators: []any{stringvalidator.OneOfCaseInsensitive("on", "off")}, // want `string property "logging" only takes "on" and "off" - make it a bool`
			},

			"tier": fwschema.StringAttribute{
				Optional:   true,
				Validators: []any{stringvalidator.OneOf("Basic", "Standard")},
			},

			"soft_delete": fwschema.StringAttribute{
				Optional:           true,
				DeprecationMessage: "use `soft_delete_enabled`",
				Validators:         []any{stringvalidator.OneOf("Enabled", "Disabled")},
			},
		},
	}
}
