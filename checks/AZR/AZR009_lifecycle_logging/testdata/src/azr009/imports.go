package azr009

import (
	"fmt"
	"log"
)

func resourceOther() *Resource {
	return &Resource{
		Delete: resourceOtherDelete,
	}
}

// Should be flagged, and with nothing else in the file using log, the import goes too.
func resourceOtherDelete(d *ResourceData, meta interface{}) error {
	id := d.Id()
	log.Printf("[DEBUG] Deleting %s", id) // want `lifecycle logging should be removed`
	log.Printf("[DEBUG] Deleted %s.", id) // want `lifecycle logging should be removed`
	return fmt.Errorf("deleting %s", id)
}
