package azs027

import "azs027/pluginsdk"

func location() *pluginsdk.Schema { return nil }

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"tags": { // want `property "tags" should come after "sku_name" - it goes last`
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"name": { // want `property "name" should come before "tags" - name, resource_group_name, location go first`
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			// a property from a function is placed like any other
			"location": location(), // want `property "location" should come before "tags"`

			"resource_group_name": { // want `property "resource_group_name" should come before "tags"`
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			// a block is its own small thing
			"rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
						},

						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

// the first three are in order among themselves too
func swapped() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"name": { // want `property "name" should come before "resource_group_name" - name, resource_group_name, location go first`
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}

type ThingResource struct{}

// in order, and a resource need not have all of them
func (ThingResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"location": location(),

		"sku_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"tags": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},
	}
}

// what is only read back is kept in order of its own
func (ThingResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"endpoint": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"location": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}
