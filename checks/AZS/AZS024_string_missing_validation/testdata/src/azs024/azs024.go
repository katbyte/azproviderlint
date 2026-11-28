package azs024

import (
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"

	fwschema "azs024/framework/schema"
	"azs024/other"
	"azs024/pluginsdk"
	"azs024/stringvalidator"
	"azs024/validation"
)

func notEmpty(any, string) ([]string, []error) { return nil, nil }

func noError(any, string) error { return nil }

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": { // want `string property "name" has no validation - add a ValidateFunc that checks what the API accepts`
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sku": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: notEmpty,
			},

			// not being empty is validation enough until the rule is strict
			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"tier": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateDiagFunc: noError,
			},

			// set or read back is still set, and nil checks nothing
			"location": { // want `string property "location" has no validation`
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},

			"zone": { // want `string property "zone" has no validation`
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: nil,
			},

			// what is only read back is not the user's to get wrong
			"endpoint": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			// the values of a list or set are checked one by one
			"zones": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{ // want `the strings in "zones" have no validation - add a ValidateFunc that checks what the API accepts`
					Type: pluginsdk.TypeString,
				},
			},

			"scopes": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: notEmpty,
				},
			},

			"addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			// the values of a map cannot be checked one by one
			"tags": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			// in a block too
			"rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"action": { // want `string property "action" has no validation`
							Type:     pluginsdk.TypeString,
							Required: true,
						},

						"priority": {
							Type:     pluginsdk.TypeInt,
							Required: true,
						},
					},
				},
			},

			// anything but a string is another rule's
			"enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"ports": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeInt,
				},
			},

			// a property on its way out is left alone
			"kind": {
				Type:       pluginsdk.TypeString,
				Optional:   true,
				Deprecated: "use `sku`",
			},

			// a schema from a function outside the provider is reported where it is used, as that is where it can be changed
			"helper_zone":     commonschema.ZoneSingleOptionalComputed(), // want `string property "helper_zone" has no validation in commonschema.ZoneSingleOptionalComputed\(\) - add a ValidateFunc that checks what the API accepts`
			"zone_optional":   commonschema.ZoneSingleOptional(),         // want `string property "zone_optional" has no validation in commonschema.ZoneSingleOptional\(\)`
			"helper_zones":    commonschema.ZonesMultipleOptional(),      // want `the strings in "helper_zones" have no validation in commonschema.ZonesMultipleOptional\(\)`
			"helper_location": commonschema.Location(),
			"name_from":       commonschema.Name(),
			"zone_computed":   commonschema.ZoneComputed(),
			"zone_from_here":  property(),

			"zone_list": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem:     commonschema.Zone(), // want `the strings in "zone_list" have no validation in commonschema.Zone\(\)`
			},

			// what the user does not set is left alone, wherever its values come from
			"zone_list_computed": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem:     commonschema.Zone(),
			},
		},
	}
}

// a function in the provider is reported where its schema is written, and not again where it is used
func property() *pluginsdk.Schema {
	return &pluginsdk.Schema{ // want `string property has no validation - add a ValidateFunc that checks what the API accepts`
		Type:     pluginsdk.TypeString,
		Optional: true,
	}
}

func list() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeSet,
		Optional: true,
		Elem: &pluginsdk.Schema{ // want `the strings in this property have no validation`
			Type: pluginsdk.TypeString,
		},
	}
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{ // want `string property "name" has no validation - add Validators that check what the API accepts`
				Required: true,
			},

			"sku": fwschema.StringAttribute{
				Optional:   true,
				Validators: []any{notEmpty},
			},

			"description": fwschema.StringAttribute{
				Optional:   true,
				Validators: []any{stringvalidator.LengthAtLeast(1)},
			},

			"tier": fwschema.StringAttribute{
				Optional:   true,
				Validators: validators(),
			},

			"zone": fwschema.StringAttribute{ // want `string property "zone" has no validation`
				Optional:   true,
				Computed:   true,
				Validators: []any{},
			},

			"endpoint": fwschema.StringAttribute{
				Computed: true,
			},

			"kind": fwschema.StringAttribute{
				Optional:           true,
				DeprecationMessage: "use `sku`",
			},

			"enabled": fwschema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func validators() []any { return nil }

// a type that only looks like a schema is not one
var _ = other.Schema{Type: 1, Required: true}
