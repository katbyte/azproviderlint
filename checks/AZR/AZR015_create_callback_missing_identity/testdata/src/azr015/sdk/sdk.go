// Package sdk is a minimal stub of the provider's internal/sdk for the analyzer tests.
package sdk

import "azr015/schema"

type ResourceMetaData struct{}

func SetIDCallback(client any, id any, d *schema.ResourceData) func() error { return nil }

func SetIDAndIdentityCallback(client any, id any, d *schema.ResourceData) func() error { return nil }

func (ResourceMetaData) SetIDCallback(id any) func() error { return nil }

func (ResourceMetaData) SetIDAndIdentityCallback(id any) func() error { return nil }

type ResourceFunc struct {
	Func func(metadata ResourceMetaData) error
}
