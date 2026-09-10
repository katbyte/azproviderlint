package azs009

import (
	"azs009/pluginsdk"
	"azs009/schema"
)

func validateString(any, string) ([]string, []error) { return nil, nil }

func suppress(string, string, string, any) bool { return false }

func flaggedTopLevel() {
	_ = &schema.Schema{
		Type:         schema.TypeString,
		Computed:     true,
		ValidateFunc: validateString, // want `ValidateFunc has no effect on a computed-only field - remove it`
	}

	_ = &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Computed: true,
		MaxItems: 1, // want `MaxItems has no effect on a computed-only field - remove it`
		MinItems: 1, // want `MinItems has no effect on a computed-only field - remove it`
		Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
	}

	_ = &schema.Schema{
		Type:             schema.TypeString,
		Computed:         true,
		Default:          "x",             // want `Default has no effect on a computed-only field - remove it`
		DiffSuppressFunc: suppress,        // want `DiffSuppressFunc has no effect on a computed-only field - remove it`
		ConflictsWith:    []string{"other"}, // want `ConflictsWith has no effect on a computed-only field - remove it`
		WriteOnly:        true,            // want `WriteOnly has no effect on a computed-only field - remove it`
	}

	_ = &schema.Schema{
		Type:       schema.TypeList,
		Computed:   true,
		ConfigMode: schema.SchemaConfigModeBlock, // want `ConfigMode: SchemaConfigModeBlock has no effect on a computed-only field - remove it`
		Elem:       &schema.Resource{Schema: map[string]*schema.Schema{}},
	}
}

func flaggedNested() {
	_ = &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Computed: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name": {
					Type:         pluginsdk.TypeString,
					Optional:     true,           // want `Optional has no effect inside a computed-only block - use Computed`
					ValidateFunc: validateString, // want `ValidateFunc has no effect inside a computed-only block - remove it`
				},
				"enabled": {
					Type:     pluginsdk.TypeBool,
					Optional: true, // want `Optional has no effect inside a computed-only block - use Computed`
					Computed: true,
				},
				"count": {
					Type:     pluginsdk.TypeInt,
					Required: true, // want `Required has no effect inside a computed-only block - use Computed`
				},
				"inner": {
					Type:     pluginsdk.TypeList,
					Computed: true,
					MaxItems: 1, // want `MaxItems has no effect inside a computed-only block - remove it`
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"deep": {
								Type:     pluginsdk.TypeString,
								Optional: true, // want `Optional has no effect inside a computed-only block - use Computed`
							},
						},
					},
				},
			},
		},
	}

	_ = &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateString, // want `ValidateFunc has no effect on the element of a computed-only field - remove it`
		},
	}
}

type typedResource struct{}

func (typedResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validateString,
		},
	}
}

func (typedResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"id": {
			Type:         pluginsdk.TypeString,
			ValidateFunc: validateString, // want `ValidateFunc has no effect in Attributes\(\), where every field is computed - remove it`
		},
		"misplaced": {
			Type:     pluginsdk.TypeString,
			Optional: true, // want `Optional has no effect in Attributes\(\), where every field is computed - use Computed`
		},
		"block": {
			Type: pluginsdk.TypeList,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"port": {
						Type:     pluginsdk.TypeInt,
						Optional: true, // want `Optional has no effect inside a computed-only block - use Computed`
					},
				},
			},
		},
	}
}

func passing(dynamic bool) {
	// optional+computed is configurable, everything is allowed
	_ = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validateString,
	}

	// an optional block may nest optional fields
	_ = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateString,
				},
			},
		},
	}

	// zero values are not set
	_ = &schema.Schema{
		Type:          schema.TypeList,
		Computed:      true,
		MaxItems:      0,
		Default:       nil,
		ConflictsWith: nil,
		Elem:          &schema.Schema{Type: schema.TypeString},
	}

	// computed under a non-constant flag is not provably computed-only
	_ = &schema.Schema{
		Type:         schema.TypeString,
		Computed:     dynamic,
		ValidateFunc: validateString,
	}

	// plain computed fields with output-side attributes are fine
	_ = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "set by the API",
	}
}
