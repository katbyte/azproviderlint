package azt005

import "azt005/pluginsdk"

func resourceLegacy() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: noop,
		Read:   noop,
		Delete: noop,
	}
}
