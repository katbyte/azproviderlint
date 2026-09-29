package azr009

import "log"

func resourceSingle() *Resource {
	return &Resource{
		Update: resourceSingleUpdate,
	}
}

// Should be flagged, and with log imported on its own line rather than in a block, the whole
// import declaration goes.
func resourceSingleUpdate(d *ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating %s", d.Id()) // want `lifecycle logging should be removed`
	return nil
}
