package azs005

import (
	fwschema "azs005/framework/schema"
	"azs005/other"
	"azs005/pluginsdk"
)

func validate(any, string) ([]string, []error) { return nil, nil }

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			// in order
			"name": {
				Description:  "The name of the Thing.",
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate,
			},

			// set or read back, in order
			"location": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},

			"sku": {
				Type:        pluginsdk.TypeString,
				Description: "The SKU of the Thing.", // want `schema field Description should come before Type`
				Optional:    true,
			},

			"zone": {
				Type:     pluginsdk.TypeString,
				Computed: true,
				// Note: O+C the API picks one when it is not set
				Optional: true, // want `schema field Optional should come before Computed`
			},

			"mode": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate,
				// the API's own default
				Default:  "Fast",
				ForceNew: true, // want `schema field ForceNew should come before ValidateFunc`
			},

			"enabled": {Optional: true, Type: pluginsdk.TypeBool}, // want `schema field Type should come before Optional`

			// what a property holds goes last, and the properties inside it are put in order with it
			"rule": {
				Type: pluginsdk.TypeList,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"priority": {
							Required: true,
							Type:     pluginsdk.TypeInt, // want `schema field Type should come before Required`
						},
					},
				},
				MaxItems: 1,
				Optional: true, // want `schema field Optional should come before Elem`
			},

			"tags": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Schema{
					ValidateFunc: validate,
					Type:         pluginsdk.TypeString, // want `schema field Type should come before ValidateFunc`
				},
			},

			// two fields on a line are reported and left for a person to move
			"size": {
				Optional: true, Type: pluginsdk.TypeInt, // want `schema field Type should come before Optional`
			},
		},
	}
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			// in order
			"name": fwschema.StringAttribute{
				Description: "The name of the Thing.",
				Required:    true,
				Validators:  []any{},
			},

			"sku": fwschema.StringAttribute{
				Optional:    true,
				Description: "The SKU of the Thing.", // want `schema field Description should come before Optional`
				Computed:    true,
			},

			"rule": fwschema.ListNestedAttribute{
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"priority": fwschema.StringAttribute{
							Validators: []any{},
							Required:   true, // want `schema field Required should come before Validators`
						},
					},
				},
				Optional: true, // want `schema field Optional should come before NestedObject`
			},
		},
		Description: "A Thing.",
	}
}

// a type that only looks like a schema is not one
var _ = other.Schema{Optional: true, Type: 1}
