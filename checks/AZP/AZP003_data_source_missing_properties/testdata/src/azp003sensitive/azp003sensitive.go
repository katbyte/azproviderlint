package azp003sensitive

import (
	"github.com/example/provider/pluginsdk"
)

// With -AZP003.ignore-sensitive, `Sensitive: true` resource properties are exempt while
// ordinary properties are still required of the data source.

type Registration struct{}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_secretive": resourceSecretive(),
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_secretive": dataSourceSecretive(), // want `data source "azurerm_secretive" is missing resource properties: "zone"`
	}
}

func resourceSecretive() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
			"zone": {Type: pluginsdk.TypeString, Optional: true},
			"primary_access_key": {
				Type:      pluginsdk.TypeString,
				Sensitive: true,
			},
			"credentials": {
				Type:      pluginsdk.TypeList,
				Sensitive: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"secret": {Type: pluginsdk.TypeString, Optional: true},
					},
				},
			},
		},
	}
}

func dataSourceSecretive() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
	}
}
