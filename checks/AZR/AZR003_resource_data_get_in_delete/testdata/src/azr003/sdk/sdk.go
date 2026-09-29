// Package sdk is a minimal stand-in for the provider's typed-resource SDK.
package sdk

type ResourceData struct{}

func (d *ResourceData) Get(key string) interface{} { return nil }
func (d *ResourceData) Id() string                 { return "" }

type ResourceMetaData struct {
	ResourceData *ResourceData
}

type ResourceFunc struct {
	Func func(metadata ResourceMetaData) error
}

// ResourceDelete is a delete handler living outside the package that registers it.
func ResourceDelete(d *ResourceData, meta interface{}) error { return nil }
