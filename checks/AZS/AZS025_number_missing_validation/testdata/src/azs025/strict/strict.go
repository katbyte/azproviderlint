// Package strict is checked with strict set, when a number only checked for not being zero or negative is reported too.
package strict

import (
	fwschema "azs025/framework/schema"
	"azs025/int64validator"
	"azs025/pluginsdk"
	"azs025/validation"
)

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"count": { // want `number property "count" is only checked for not being zero or negative - add a ValidateFunc that checks what the API accepts`
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			"retries": { // want `number property "retries" is only checked for not being zero or negative`
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"ratio": { // want `number property "ratio" is only checked for not being zero or negative`
				Type:         pluginsdk.TypeFloat,
				Optional:     true,
				ValidateFunc: validation.FloatAtLeast(0),
			},

			"instances": { // want `number property "instances" is only checked for not being zero or negative`
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntPositive,
			},

			"port": { // want `number property "port" is only checked for not being zero or negative`
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntNotInSlice([]int{0}),
			},

			"size": { // want `number property "size" is only checked for not being zero or negative`
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"ports": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{ // want `the numbers in "ports" are only checked for not being zero or negative`
					Type:         pluginsdk.TypeInt,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},

			// with no validation at all, it is reported as it always is
			"unchecked": { // want `number property "unchecked" has no validation`
				Type:     pluginsdk.TypeInt,
				Optional: true,
			},

			// a lowest value above 1 is one the API asks for, and so is a range
			"nodes": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(3),
			},

			"days": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 365),
			},

			"weight": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntNotInSlice([]int{0, 100}),
			},
		},
	}
}

func attributes() map[string]fwschema.Attribute {
	return map[string]fwschema.Attribute{
		"count": fwschema.Int64Attribute{ // want `number property "count" is only checked for not being zero or negative - add Validators that check what the API accepts`
			Optional:   true,
			Validators: []any{int64validator.AtLeast(1)},
		},

		"port": fwschema.Int64Attribute{ // want `number property "port" is only checked for not being zero or negative`
			Optional:   true,
			Validators: []any{int64validator.NoneOf(0)},
		},

		// a lowest and a highest value together is a range
		"days": fwschema.Int64Attribute{
			Optional:   true,
			Validators: []any{int64validator.AtLeast(1), int64validator.AtMost(365)},
		},
	}
}
