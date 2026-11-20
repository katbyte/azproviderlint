package azt005

import "azt005/pluginsdk"

func dataSourceFixed() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: noop,
	}
}
