// Package sdk is a minimal stub of the provider's internal/sdk for the analyzer tests.
package sdk

import (
	"azr012/clients"
	"azr012/tf"
)

type ResourceMetaData struct {
	Client *clients.Client
}

// the function that builds the error for the other is not a create
func (rmd ResourceMetaData) ResourceRequiresImport(resourceName string, id interface{ ID() string }) error {
	return tf.ImportAsExistsError(resourceName, id.ID())
}
