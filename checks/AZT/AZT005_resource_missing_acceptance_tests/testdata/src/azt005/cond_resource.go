package azt005

import "azt005/pluginsdk"

func resourceCond() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: noop,
		Read:   noop,
		Delete: noop,
	}
}
