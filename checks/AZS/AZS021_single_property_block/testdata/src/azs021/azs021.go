package azs021

import (
	"azs021/other"
	"azs021/pluginsdk"
)

func resource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"certificate": { // want `block "certificate" can be given once and holds only "thumbprint" - move it out of the block, or say why the block is kept in a comment above it`
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"thumbprint": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"identity": &pluginsdk.Schema{ // want `block "identity" can be given once and holds only "type"`
				Type:     pluginsdk.TypeSet,
				Required: true,
				MaxItems: 1,

				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"type": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			// two properties are worth a block
			"sku": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},

						"capacity": {
							Type:     pluginsdk.TypeInt,
							Optional: true,
						},
					},
				},
			},

			// the service team has said more settings are coming to this block
			"backup": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"enabled": {
							Type:     pluginsdk.TypeBool,
							Required: true,
						},
					},
				},
			},

			"network": { // want `block "network" can be given once and holds only "subnet_id"`
				Type:     pluginsdk.TypeList,
				Optional: true,
				// a comment inside the block is about the field under it, not about why the block is kept
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"subnet_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			// lintignore:XS003
			"route": { // want `block "route" can be given once and holds only "prefix"`
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"prefix": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"storage": { // want `block "storage" can be given once and holds only "account_id"`
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"account_id": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			}, // the comment that ends a line is about that line
			"logging": { // want `block "logging" can be given once and holds only "level"`
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1, // nor does this one say why
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						// a comment on the property is not one on the block
						"level": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			// a block that can be given more than once, or any number of times, is not a wrapper
			"rule": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 2,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"tag": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			// a list of one string is already flat
			"zones": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"identity_from_function": { // want `block "identity_from_function" can be given once and holds one property - move it out of the block`
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     identity(),
			},

			"backup_from_function": { // want `block "backup_from_function" can be given once and holds one property - move it out of the block`
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: backup(),
				},
			},

			"network_from_function": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     network(),
			},

			// one handed back by a function is kept with a comment like any other
			"logging_from_function": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     identity(),
			},

			// none, or ones that cannot be read, are not counted
			"encryption": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     encryption(),
			},

			"replication": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: replication(),
				},
			},
		},
	}
}

func encryption() *pluginsdk.Resource { return &pluginsdk.Resource{} }

func replication() map[string]*pluginsdk.Schema { return nil }

func identity() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"type": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
		},
	}
}

func backup() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"retention_days": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},
	}
}

func network() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"subnet_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"private_ip_address": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},
		},
	}
}

func scaling() *pluginsdk.Schema {
	return &pluginsdk.Schema{ // want `block can be given once and holds only "capacity" - move it out of the block, or say why the block is kept in a comment above it`
		Type:     pluginsdk.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"capacity": {
					Type:     pluginsdk.TypeInt,
					Required: true,
				},
			},
		},
	}
}

// a type that only looks like a schema is not one
var _ = other.Schema{
	MaxItems: 1,
	Elem: &other.Resource{
		Schema: map[string]*other.Schema{
			"name": {},
		},
	},
}
