package azr015

import (
	"azr015/schema"
	"azr015/sdk"
)

type thingId struct{}

func poll(callback func() error) error { return nil }

func resourceThing() *schema.Resource {
	return &schema.Resource{
		Create:   resourceThingCreate,
		Update:   resourceThingUpdate,
		Identity: &schema.ResourceIdentity{},
	}
}

func resourceThingCreate(d *schema.ResourceData, meta any) error {
	id := thingId{}

	return poll(sdk.SetIDCallback(meta, &id, d)) // want `this resource has an identity, and SetIDCallback sets only the ID - use SetIDAndIdentityCallback`
}

// by an update both are already saved
func resourceThingUpdate(d *schema.ResourceData, meta any) error {
	id := thingId{}

	return poll(sdk.SetIDCallback(meta, &id, d))
}

func resourceOther() *schema.Resource {
	return &schema.Resource{
		Create:   resourceOtherCreate,
		Identity: &schema.ResourceIdentity{},
	}
}

func resourceOtherCreate(d *schema.ResourceData, meta any) error {
	id := thingId{}

	return poll(sdk.SetIDAndIdentityCallback(meta, &id, d))
}

// a resource with no identity has none to set
func resourcePlain() *schema.Resource {
	return &schema.Resource{
		Create: resourcePlainCreate,
	}
}

func resourcePlainCreate(d *schema.ResourceData, meta any) error {
	id := thingId{}

	return poll(sdk.SetIDCallback(meta, &id, d))
}

type ThingResource struct{}

func (ThingResource) Identity() any { return &thingId{} }

func (r ThingResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(metadata sdk.ResourceMetaData) error {
			id := thingId{}

			return poll(metadata.SetIDCallback(&id)) // want `this resource has an identity, and SetIDCallback sets only the ID - use SetIDAndIdentityCallback`
		},
	}
}

func (r ThingResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(metadata sdk.ResourceMetaData) error {
			id := thingId{}

			return poll(metadata.SetIDCallback(&id))
		},
	}
}

type PlainResource struct{}

func (r PlainResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(metadata sdk.ResourceMetaData) error {
			id := thingId{}

			return poll(metadata.SetIDCallback(&id))
		},
	}
}
