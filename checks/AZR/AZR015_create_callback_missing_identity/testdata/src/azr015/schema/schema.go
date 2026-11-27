// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ResourceData struct{}

type ResourceIdentity struct{}

type Resource struct {
	Create   func(d *ResourceData, meta any) error
	Update   func(d *ResourceData, meta any) error
	Identity *ResourceIdentity
}
