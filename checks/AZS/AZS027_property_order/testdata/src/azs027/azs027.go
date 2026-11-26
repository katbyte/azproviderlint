package azs027

import "azs027/pluginsdk"

func location() *pluginsdk.Schema { return nil }

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: nil,

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

// what the resource sits in is part of its ID, so it may sit among the first, and what is only read back is listed after tags
func child() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"lab_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"workspace_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"location": location(),

		"tags": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"endpoint": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

// what the user can set as well as read back is not read-only, and a property from a function cannot be read
func afterTags() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"tags": { // want `property "tags" should come after "zone" - it goes last`
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"endpoint": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"zone": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Computed: true,
		},
	}
}

// what a block holds is not a resource, however it is built: a resource says how it is read
func rule() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"action": {
					Type:     pluginsdk.TypeString,
					Required: true,
				},

				"name": {
					Type:     pluginsdk.TypeString,
					Required: true,
				},
			},
		},
	}
}

func header() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"action": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
		},
	}
}

// a function with DataSource in its name builds one
func DataSourceThingSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"id": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}

// a property that must be given and cannot change is taken as part of the ID, whatever it is called, and a name that is only read back is not
// the one the user gives
func identified() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"identifier": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"description": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"name": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},
	}
}

// one that can change is an ordinary property
func ordinary() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"sku": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"location": { // want `property "location" should come before "sku" - name, resource_group_name, location go first`
			Type:     pluginsdk.TypeString,
			Required: true,
		},
	}
}
