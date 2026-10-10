package azt008

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type Registration struct{}

const probeName = "azurerm_lb_probe"

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_thing": resourceThing(),
		probeName:       resourceLbProbe(), // named by a constant
	}

	// registered behind a feature flag
	if featureFlag() {
		resources["azurerm_flagged"] = resourceFlagged()
	}

	return resources
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_thing": dataSourceThing(),
	}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		TypedResource{},
	}
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		&TypedDataSource{},
	}
}

func featureFlag() bool {
	return false
}
