package azt005disable

import "azt005disable/pluginsdk"

func resourceUntested() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: noop,
		Read:   noop,
		Update: noop,
		Delete: noop,
	}
}
