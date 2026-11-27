// Package schema is a minimal stub of the plugin SDK's helper/schema for the analyzer tests.
package schema

type ResourceData struct{}

func (*ResourceData) HasChange(key string) bool { return false }

func (*ResourceData) HasChanges(keys ...string) bool { return false }
