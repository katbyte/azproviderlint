// Package strict is checked with strict set, when a string only checked for not being empty is reported too.
package strict

import (
	fwschema "azs024/framework/schema"
	"azs024/pluginsdk"
	"azs024/stringvalidator"
	"azs024/validation"
)

var StringIsNotEmpty = validation.StringIsNotEmpty

func noError(any, string) error { return nil }

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": { // want `string property "name" is only checked for not being empty - add a ValidateFunc that checks what the API accepts`
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			// a provider's own copy of it says no more
			"sku": { // want `string property "sku" is only checked for not being empty`
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: StringIsNotEmpty,
			},

			"zones": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{ // want `the strings in "zones" are only checked for not being empty - add a ValidateFunc that checks what the API accepts`
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},

			// none at all is still reported as that
			"location": { // want `string property "location" has no validation`
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			// anything more is validation
			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 80),
			},

			"tier": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsNotEmpty,
				ValidateDiagFunc: noError,
			},

			// what is only read back is still not the user's to get wrong
			"endpoint": {
				Type:         pluginsdk.TypeString,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{ // want `string property "name" is only checked for not being empty - add Validators that check what the API accepts`
				Required:   true,
				Validators: []any{stringvalidator.LengthAtLeast(1)},
			},

			"sku": fwschema.StringAttribute{
				Optional:   true,
				Validators: []any{stringvalidator.LengthAtLeast(3)},
			},

			"tier": fwschema.StringAttribute{
				Optional:   true,
				Validators: []any{stringvalidator.LengthAtLeast(1), stringvalidator.LengthBetween(1, 80)},
			},

			"location": fwschema.StringAttribute{ // want `string property "location" has no validation`
				Optional: true,
			},
		},
	}
}
