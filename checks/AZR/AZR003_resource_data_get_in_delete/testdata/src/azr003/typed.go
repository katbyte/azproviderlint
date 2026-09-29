package azr003

import "azr003/sdk"

type SDKTypedResource struct{}

// Should be flagged: the typed Delete method with the real `sdk.ResourceFunc` result type.
func (r SDKTypedResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(metadata sdk.ResourceMetaData) error {
			name := metadata.ResourceData.Get("name") // want `ResourceData\.Get should not be used within a Delete function as it does not work as expected during deletion`
			_ = name
			_ = metadata.ResourceData.Id()
			return nil
		},
	}
}

// Should NOT be flagged: a Delete method that does not return a ResourceFunc is not a
// lifecycle step, whatever it reads.
type deleter struct{}

func (deleter) Delete(metadata sdk.ResourceMetaData) error {
	_ = metadata.ResourceData.Get("name")
	return nil
}

func (deleter) DeleteContext(metadata sdk.ResourceMetaData) {
	_ = metadata.ResourceData.Get("name")
}

// Should be flagged: the read goes through a local alias of ResourceData.
type AliasedTypedResource struct{}

func (r AliasedTypedResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(metadata sdk.ResourceMetaData) error {
			d := metadata.ResourceData
			_ = d.Get("name") // want `d\.Get should not be used within a Delete function as it does not work as expected during deletion`
			rd := metadata.ResourceData
			_ = rd.Get("name") // want `rd\.Get should not be used within a Delete function as it does not work as expected during deletion`
			other := d.Id()
			_ = other
			return nil
		},
	}
}

type sdkResource struct {
	Delete func(d *sdk.ResourceData, meta interface{}) error
}

// Should NOT be flagged: a handler registered from another package has no declaration here
// to inspect.
func resourceExternalDelete() *sdkResource {
	return &sdkResource{
		Delete: sdk.ResourceDelete,
	}
}

// Should NOT be flagged: a registered Delete function with no parameters has no data to read.
func resourceNoParams() *Resource {
	return &Resource{
		Delete: resourceNoParamsDelete,
	}
}

func resourceNoParamsDelete(*ResourceData, interface{}) error {
	return nil
}

// Should NOT be flagged: Get on something other than the data parameter, and a method other
// than Get on the data parameter.
func resourceOtherGet() *Resource {
	return &Resource{
		Delete: resourceOtherGetDelete,
	}
}

func resourceOtherGetDelete(d *ResourceData, meta interface{}) error {
	other := &ResourceData{}
	_ = other.Get("name")
	_ = d.Id()
	return nil
}

// Should NOT be flagged: a Delete method returning anything but a plain (sdk.)ResourceFunc —
// a pointer to one here — is not recognised as a lifecycle step.
type PointerResultResource struct{}

func (r PointerResultResource) Delete() *sdk.ResourceFunc {
	return &sdk.ResourceFunc{
		Func: func(metadata sdk.ResourceMetaData) error {
			_ = metadata.ResourceData.Get("name")
			return nil
		},
	}
}

// Should NOT be flagged: a Delete registered as a function literal has no declaration to
// inspect.
func resourceInlineDelete() *Resource {
	return &Resource{
		Delete: func(d *ResourceData, meta interface{}) error {
			_ = d.Get("name")
			return nil
		},
	}
}

// Should be flagged: a method registered under Delete: is an untyped handler like any other.
type methodResource struct{}

func (r *methodResource) resource() *Resource {
	return &Resource{
		Delete: r.delete,
	}
}

func (r *methodResource) delete(d *ResourceData, meta interface{}) error {
	_ = d.Get("name") // want `d\.Get should not be used within a Delete function as it does not work as expected during deletion`
	return nil
}
