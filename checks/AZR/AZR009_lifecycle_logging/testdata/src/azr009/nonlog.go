package azr009

import (
	"fmt"
	"log"
)

func resourceOtherLoggers() *Resource {
	return &Resource{
		Create: resourceOtherLoggersCreate,
	}
}

type wrapper struct{ logger *log.Logger }

// Should NOT be flagged: narration through anything but the log package or a metadata.Logger
// method, and a log call with no message.
func resourceOtherLoggersCreate(d *ResourceData, meta interface{}) error {
	id := d.Id()
	w := wrapper{logger: log.Default()}
	fmt.Printf("Creating %s\n", id)
	println("Creating", id)
	w.logger.Printf("Creating %s", id)
	log.Println()
	log.Printf("[DEBUG] Creating %s", id) // want `lifecycle logging should be removed`
	return nil
}

// Should NOT be flagged: a Logger method outside the narration set.
func (r ThingResource) Delete() ResourceFunc {
	return ResourceFunc{
		Func: func(metadata ResourceMetaData) error {
			metadata.Logger.Error("Deleting %s")
			return nil
		},
	}
}
