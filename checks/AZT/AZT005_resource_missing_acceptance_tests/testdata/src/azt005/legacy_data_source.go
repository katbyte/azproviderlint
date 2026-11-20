package azt005

import "azt005/pluginsdk"

func dataSourceLegacy() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: noop,
	}
}
