package azs023

import (
	fwschema "azs023/framework/schema"
	"azs023/other"
	"azs023/pluginsdk"
)

const zoneRedundant = "is_zone_redundant"

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"is_storage_enabled": { // want `property "is_storage_enabled" starts with is_ - name it "storage_enabled"`
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"storage_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			// whatever its type, and wherever its schema comes from
			"is_default": { // want `property "is_default" starts with is_ - name it "default"`
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			zoneRedundant: zones(), // want `property "is_zone_redundant" starts with is_ - name it "zone_redundant"`

			// in a block too
			"rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"is_active": { // want `property "is_active" starts with is_ - name it "active"`
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},
					},
				},
			},

			// a name that only starts with the same letters
			"issuer": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"island": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"this_is_fine": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"is_": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},
		},
	}
}

func zones() *pluginsdk.Schema { return nil }

// a property added to the schema later is named there
func arguments(legacy bool) map[string]*pluginsdk.Schema {
	out := map[string]*pluginsdk.Schema{}
	if legacy {
		out["is_legacy"] = &pluginsdk.Schema{ // want `property "is_legacy" starts with is_ - name it "legacy"`
			Type:     pluginsdk.TypeBool,
			Optional: true,
		}
	}

	out["legacy_enabled"] = &pluginsdk.Schema{
		Type:     pluginsdk.TypeBool,
		Optional: true,
	}

	return out
}

func framework() fwschema.Schema {
	return fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"is_public": fwschema.BoolAttribute{ // want `property "is_public" starts with is_ - name it "public"`
				Optional: true,
			},

			"public": fwschema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]fwschema.Block{
			"is_rule": fwschema.ListNestedBlock{ // want `property "is_rule" starts with is_ - name it "rule"`
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"is_active": fwschema.BoolAttribute{ // want `property "is_active" starts with is_ - name it "active"`
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// a map of anything else is not a schema, and reading a name is not giving one
func unrelated(d map[string]any) {
	_ = map[string]string{"is_enabled": "true"}
	_ = map[string]*other.Schema{"is_enabled": {}}
	_ = d["is_enabled"]

	d["is_enabled"] = true
}
