package azs025

import (
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"

	fwschema "azs025/framework/schema"
	"azs025/pluginsdk"
	"azs025/validation"
)

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"count": { // want `number property "count" has no validation - add a ValidateFunc that checks what the API accepts`
				Type:     pluginsdk.TypeInt,
				Required: true,
			},

			"ratio": { // want `number property "ratio" has no validation`
				Type:     pluginsdk.TypeFloat,
				Optional: true,
				Computed: true,
			},

			"instances": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},

			// without strict, any check counts, even one that only says it is not negative
			"threshold": {
				Type:         pluginsdk.TypeFloat,
				Optional:     true,
				ValidateFunc: validation.FloatAtLeast(0),
			},

			// what the user does not set is left alone
			"used": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"ports": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{ // want `the numbers in "ports" have no validation`
					Type: pluginsdk.TypeInt,
				},
			},

			"weights": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeInt,
					ValidateFunc: validation.IntBetween(0, 100),
				},
			},

			// the values of a map cannot be checked one by one
			"limits": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeInt,
				},
			},

			// a string is not this rule's
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			// a property on its way out is left alone
			"size": {
				Type:       pluginsdk.TypeInt,
				Optional:   true,
				Deprecated: "use `capacity`",
			},

			// a schema from a function outside the provider is reported where it is used, as that is where it can be changed
			"capacity":        commonschema.Capacity(), // want `number property "capacity" has no validation in commonschema.Capacity\(\) - add a ValidateFunc that checks what the API accepts`
			"helper_count":    commonschema.Instances(),
			"helper_name":     commonschema.Name(),
			"count_from_here": property(),

			"port_list": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem:     commonschema.Port(), // want `the numbers in "port_list" have no validation in commonschema.Port\(\)`
			},
		},
	}
}

// a function in the provider is reported where its schema is written, and not again where it is used
func property() *pluginsdk.Schema {
	return &pluginsdk.Schema{ // want `number property has no validation - add a ValidateFunc that checks what the API accepts`
		Type:     pluginsdk.TypeInt,
		Optional: true,
	}
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"count": fwschema.Int64Attribute{ // want `number property "count" has no validation - add Validators that check what the API accepts`
				Required: true,
			},

			"ratio": fwschema.Float64Attribute{ // want `number property "ratio" has no validation - add Validators that check`
				Optional:   true,
				Validators: []any{},
			},

			"instances": fwschema.Int64Attribute{
				Optional:   true,
				Validators: []any{validation.IntBetween(1, 10)},
			},

			"used": fwschema.Int64Attribute{
				Computed: true,
			},
		},
	}
}
