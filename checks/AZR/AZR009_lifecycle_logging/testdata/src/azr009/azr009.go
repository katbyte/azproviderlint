package azr009

import (
	"log"
)

type ResourceData struct{}

func (d *ResourceData) Id() string { return "" }

type Resource struct {
	Create func(d *ResourceData, meta interface{}) error
	Read   func(d *ResourceData, meta interface{}) error
	Update func(d *ResourceData, meta interface{}) error
	Delete func(d *ResourceData, meta interface{}) error
}

func resourceThing() *Resource {
	return &Resource{
		Create: resourceThingCreate,
		Read:   resourceThingRead,
		Update: resourceThingUpdate,
		Delete: resourceThingDelete,
	}
}

// Should be flagged: narration of the step, with and without a level tag.
func resourceThingCreate(d *ResourceData, meta interface{}) error {
	log.Printf("[INFO] preparing arguments for Azure Thing creation.") // want `lifecycle logging should be removed`
	id := d.Id()
	log.Printf("Creating %s", id) // want `lifecycle logging should be removed`
	return nil
}

// Should be flagged: retrieving; should NOT be flagged: removing from state.
func resourceThingRead(d *ResourceData, meta interface{}) error {
	id := d.Id()
	log.Printf("[DEBUG] Retrieving %s..", id) // want `lifecycle logging should be removed`
	if id == "" {
		log.Printf("[DEBUG] %s was not found - removing from state!", id)
		return nil
	}
	return nil
}

// Should NOT be flagged: an ID rewrite records a state migration; a sub-step IS flagged.
func resourceThingUpdate(d *ResourceData, meta interface{}) error {
	id := d.Id()
	log.Printf("[DEBUG] Updating ID from %s to %s", id, id)
	log.Printf("[DEBUG] Updating the network settings for %s..", id) // want `lifecycle logging should be removed`
	return nil
}

// Should NOT be flagged: polling messages and non-literal formats stay.
func resourceThingDelete(d *ResourceData, meta interface{}) error {
	id := d.Id()
	format := "[DEBUG] Deleting %s"
	log.Printf(format, id)
	log.Printf("[DEBUG] Waiting for %s to be deleted..", id)
	log.Printf("[DEBUG] deleting disks for %s - model was nil, skipping", id)
	return nil
}

// Should NOT be flagged: not a registered lifecycle function.
func helperThing(d *ResourceData) {
	log.Printf("[DEBUG] Creating %s", d.Id())
}

// typed SDK style resources

type Logger struct{}

func (Logger) Info(msg string)                          {}
func (Logger) Infof(format string, args ...interface{}) {}

type ResourceMetaData struct {
	Logger Logger
}

type ResourceFunc struct {
	Func func(metadata ResourceMetaData) error
}

type ThingResource struct{}

// Should be flagged: both the format and the plain message forms.
func (r ThingResource) Create() ResourceFunc {
	return ResourceFunc{
		Func: func(metadata ResourceMetaData) error {
			metadata.Logger.Info("Decoding state..")   // want `lifecycle logging should be removed`
			metadata.Logger.Infof("creating %s", "id") // want `lifecycle logging should be removed`
			return nil
		},
	}
}

// Should NOT be flagged: state removal, and a method that is not a lifecycle step.
func (r ThingResource) Read() ResourceFunc {
	return ResourceFunc{
		Func: func(metadata ResourceMetaData) error {
			metadata.Logger.Infof("%s was not found - removing from state!", "id")
			return nil
		},
	}
}

func (r ThingResource) Arguments() ResourceFunc {
	return ResourceFunc{
		Func: func(metadata ResourceMetaData) error {
			metadata.Logger.Infof("reading %s", "id")
			return nil
		},
	}
}
