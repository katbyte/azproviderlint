package azs024

import (
	fwschema "azs024/framework/schema"
	"azs024/other"
	"azs024/pluginsdk"
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
		},
	}
}

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
