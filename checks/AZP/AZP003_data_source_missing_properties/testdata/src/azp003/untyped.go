package azp003

import (
	"github.com/example/provider/pluginsdk"
)

// azurerm_complete: the data source covers every resource property, including the nested
// block property, so nothing is reported.

func resourceComplete() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name":                {Type: pluginsdk.TypeString, Required: true},
			"resource_group_name": {Type: pluginsdk.TypeString, Required: true},
			"settings": {
				Type: pluginsdk.TypeList,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"enabled": {Type: pluginsdk.TypeString, Optional: true},
					},
				},
			},
		},
	}
}

func dataSourceComplete() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name":                {Type: pluginsdk.TypeString, Required: true},
			"resource_group_name": {Type: pluginsdk.TypeString, Required: true},
			"settings": {
				Type: pluginsdk.TypeList,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"enabled": {Type: pluginsdk.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

// azurerm_incomplete: the resource has a top-level "zone" and a nested "retention_days" the
// data source exposes nowhere, plus a whole "networking" block the data source lacks — the
// block is reported once, without its children.

func resourceIncomplete() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name":      {Type: pluginsdk.TypeString, Required: true},
			"zone":      {Type: pluginsdk.TypeString, Optional: true},
			"secret_wo": {Type: pluginsdk.TypeString, Optional: true},
			"api_token": {Type: pluginsdk.TypeString, Optional: true, WriteOnly: true},
			"backup": {
				Type: pluginsdk.TypeList,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"retention_days": {Type: pluginsdk.TypeString, Optional: true},
					},
				},
			},
			"networking": {
				Type: pluginsdk.TypeList,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"subnet_id":  {Type: pluginsdk.TypeString, Optional: true},
						"ip_version": {Type: pluginsdk.TypeString, Optional: true},
					},
				},
			},
		},
	}
}

func dataSourceIncomplete() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name":   {Type: pluginsdk.TypeString, Required: true},
			"backup": {Type: pluginsdk.TypeList, Computed: true},
		},
	}
}

// azurerm_via_helper: both schemas are built through a shared helper function, which the
// recursive collection follows.

func sharedNameSchema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name":     {Type: pluginsdk.TypeString, Required: true},
		"location": {Type: pluginsdk.TypeString, Optional: true},
	}
}

func resourceViaHelper() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: resourceViaHelperSchema(),
	}
}

func resourceViaHelperSchema() map[string]*pluginsdk.Schema {
	s := sharedNameSchema()
	s["tier"] = &pluginsdk.Schema{Type: pluginsdk.TypeString, Optional: true}
	return s
}

func dataSourceViaHelper() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name":     {Type: pluginsdk.TypeString, Required: true},
			"location": {Type: pluginsdk.TypeString, Computed: true},
			"tier":     {Type: pluginsdk.TypeString, Computed: true},
		},
	}
}

// azurerm_dynamic_keys: the data source builds a key dynamically, so the pair is skipped even
// though the resource has properties the literal keys do not cover.

func resourceDynamicKeys() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name":    {Type: pluginsdk.TypeString, Required: true},
			"special": {Type: pluginsdk.TypeString, Optional: true},
		},
	}
}

func dynamicKey() string {
	return "name"
}

func dataSourceDynamicKeys() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			dynamicKey(): {Type: pluginsdk.TypeString, Required: true},
		},
	}
}

// azurerm_no_ds has no data source at all — that is AZP002's job, not this check's.

func resourceNoDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
	}
}
