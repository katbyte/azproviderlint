package azp002unresolved

import (
	"github.com/example/provider/pluginsdk"
)

// A data source registered under a key that cannot be resolved to a constant string skips
// the whole package: azurerm_would_report has no data source, but is not reported.

type Registration struct{}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_would_report": untypedResource(),
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	dataSources := map[string]*pluginsdk.Resource{
		"azurerm_other": untypedResource(),
	}
	dataSources[dynamicName()] = untypedResource()
	return dataSources
}

func untypedResource() *pluginsdk.Resource {
	return &pluginsdk.Resource{}
}

func dynamicName() string {
	return "azurerm_would_report"
}
