package azt005

import "azt005/pluginsdk"

func dataSourceOrphan() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: noop,
	}
}
