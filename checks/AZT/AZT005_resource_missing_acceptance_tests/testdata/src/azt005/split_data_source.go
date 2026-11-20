package azt005

import "azt005/pluginsdk"

func dataSourceSplit() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: noop,
	}
}
