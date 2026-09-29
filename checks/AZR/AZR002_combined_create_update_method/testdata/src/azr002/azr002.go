package azr002

type Resource struct {
	Create func() error
	Read   func() error
	Update func() error
	Delete func() error
}

// Should be flagged: combined CreateUpdate registered as Create
func resourceBadThing() *Resource {
	return &Resource{
		Create: resourceBadThingCreateUpdate, // want `new resources should use separate Create and Update methods instead of a combined CreateUpdate method`
		Read:   resourceBadThingRead,
		Update: resourceBadThingCreateUpdate,
		Delete: resourceBadThingDelete,
	}
}

// Should NOT be flagged: separate Create and Update methods
func resourceGoodThing() *Resource {
	return &Resource{
		Create: resourceGoodThingCreate,
		Read:   resourceGoodThingRead,
		Update: resourceGoodThingUpdate,
		Delete: resourceGoodThingDelete,
	}
}

func resourceBadThingCreateUpdate() error { return nil }
func resourceBadThingRead() error         { return nil }
func resourceBadThingDelete() error       { return nil }

func resourceGoodThingCreate() error { return nil }
func resourceGoodThingRead() error   { return nil }
func resourceGoodThingUpdate() error { return nil }
func resourceGoodThingDelete() error { return nil }

var handlers = struct {
	ThingCreateUpdate func() error
	ThingCreate       func() error
}{}

// Should be flagged: the combined method is referenced through a selector
func resourceSelectorThing() *Resource {
	return &Resource{
		Create: handlers.ThingCreateUpdate, // want `new resources should use separate Create and Update methods instead of a combined CreateUpdate method`
		Update: handlers.ThingCreateUpdate,
	}
}

// Should NOT be flagged: a selector to a plain Create method
func resourceSelectorGood() *Resource {
	return &Resource{
		Create: handlers.ThingCreate,
	}
}

// Should NOT be flagged: a function literal has no name to match
func resourceLiteralCreate() *Resource {
	return &Resource{
		Create: func() error { return nil },
	}
}

// Should NOT be flagged: a string key is not the Create field
var byName = map[string]func() error{
	"Create": resourceBadThingCreateUpdate,
}
