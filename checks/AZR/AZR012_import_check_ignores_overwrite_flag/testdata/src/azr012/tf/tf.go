// Package tf is a minimal stub of the provider's internal/tf for the analyzer tests.
package tf

import "fmt"

func ImportAsExistsError(resourceName, id string) error {
	return fmt.Errorf("a resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for %q for more information", id, resourceName)
}
