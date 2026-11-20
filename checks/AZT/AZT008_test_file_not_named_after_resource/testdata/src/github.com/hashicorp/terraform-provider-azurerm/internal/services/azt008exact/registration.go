package azt008exact

import "github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"

type Registration struct{}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_thing": resourceThing(),
	}
}
