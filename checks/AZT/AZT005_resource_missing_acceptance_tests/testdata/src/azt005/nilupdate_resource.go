package azt005

import "azt005/pluginsdk"

func resourceNilUpdate() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: noop,
		Read:   noop,
		Update: nil,
		Delete: noop,
	}
}
