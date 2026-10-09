package azt005disable

import "azt005disable/pluginsdk"

type Registration struct{}

// With requiresImport, complete and update disabled only basic is asked for
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_untested": resourceUntested(), // want `resource "azurerm_untested" is missing acceptance tests: TestAccUntested_basic`
		"azurerm_partial":  resourcePartial(),  // has basic, so nothing to ask for
	}
}

func noop() error {
	return nil
}
