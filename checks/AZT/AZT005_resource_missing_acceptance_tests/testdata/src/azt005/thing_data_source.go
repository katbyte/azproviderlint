package azt005

import "azt005/pluginsdk"

func dataSourceThing() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: noop,
	}
}
