// Package helpers hands back properties and blocks, and what each gives is recorded for where they are used.
package helpers

import (
	fwschema "framework/schema"
	"schema"
)

func validate(any, string) ([]string, []error) { return nil, nil }

func StringIsNotEmpty(any, string) ([]string, []error) { return nil, nil }

func IntPositive(any, string) ([]string, []error) { return nil, nil }

func IntAtLeast(int) func(any, string) ([]string, []error) { return nil }

func Name() *schema.Schema { // want Name:"gives\\(TypeString,required\\)"
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
}

// one handed on from another gives what that one does
func NameAgain() *schema.Schema { // want NameAgain:"gives\\(TypeString,required\\)"
	return Name()
}

func Count() *schema.Schema { // want Count:"gives\\(TypeInt,optional,validated,inputs\\)"
	return &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: validate,
	}
}

func Zones() *schema.Schema { // want Zones:"gives\\(TypeSet,optional,elem=TypeString\\)"
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func Port() *schema.Schema { // want Port:"gives\\(TypeInt\\)"
	return &schema.Schema{
		Type: schema.TypeInt,
	}
}

func Ports() *schema.Schema { // want Ports:"gives\\(TypeList,required,elem=TypeInt\\)"
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem:     Port(),
	}
}

func Rules() *schema.Schema { // want Rules:"gives\\(TypeList,computed,elem=Resource\\)"
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem:     Block(),
	}
}

func Ratio() fwschema.Float64Attribute { // want Ratio:"gives\\(Float64Attribute,optional,framework\\)"
	return fwschema.Float64Attribute{
		Optional: true,
	}
}

func NotEmpty() *schema.Schema { // want NotEmpty:"gives\\(TypeString,required,validated,loose,inputs\\)"
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: StringIsNotEmpty,
	}
}

// a number only checked for not being zero or negative says next to nothing, as does a list of them; a lowest value above 1 says something
func Positive() *schema.Schema { // want Positive:"gives\\(TypeInt,optional,validated,loose,inputs\\)"
	return &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: IntPositive,
	}
}

func Sizes() *schema.Schema { // want Sizes:"gives\\(TypeList,optional,elem-validated,elem-loose,elem=TypeInt\\)"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type:         schema.TypeInt,
			ValidateFunc: IntAtLeast(1),
		},
	}
}

func Nodes() *schema.Schema { // want Nodes:"gives\\(TypeInt,optional,validated,inputs\\)"
	return &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: IntAtLeast(3),
	}
}

func Flag() *schema.Schema { // want Flag:"gives\\(TypeBool,optional,default,inputs\\)"
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	}
}

func Old() *schema.Schema { // want Old:"gives\\(TypeString,optional,deprecated\\)"
	return &schema.Schema{
		Type:       schema.TypeString,
		Optional:   true,
		Deprecated: "gone",
	}
}

// a block's properties, written in a Resource or a map of them, or from another function
func Block() *schema.Resource { // want Block:"holds\\(\\|,known,settable\\)"
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func Properties() map[string]*schema.Schema { // want Properties:"holds\\(\\|\\|,known,stops,settable\\)"
	return map[string]*schema.Schema{
		"name": Name(),
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func ReadOnly() *schema.Resource { // want ReadOnly:"holds\\(\\|,known\\)"
	return &schema.Resource{
		Schema: readOnly(),
	}
}

func readOnly() map[string]*schema.Schema { // want readOnly:"holds\\(\\|,known\\)"
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// what a property's schema is cannot be read when it is not written out or handed back by a function that says
func Unknown(given *schema.Schema) map[string]*schema.Schema { // want Unknown:"holds\\(\\|\\)"
	return map[string]*schema.Schema{
		"given": given,
	}
}
