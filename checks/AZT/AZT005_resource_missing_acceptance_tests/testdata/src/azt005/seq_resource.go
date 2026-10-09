package azt005

import "azt005/pluginsdk"

func resourceSeq() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: noop,
		Read:   noop,
		Update: noop,
		Delete: noop,
	}
}
