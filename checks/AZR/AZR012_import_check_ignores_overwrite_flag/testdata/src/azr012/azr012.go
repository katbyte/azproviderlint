package azr012

import (
	"azr012/clients"
	"azr012/sdk"
	"azr012/tf"
)

type thingId struct{}

func (thingId) ID() string { return "" }

func exists(id thingId) bool { return false }

func resourceThingCreate(meta any) error {
	id := thingId{}
	if exists(id) {
		return tf.ImportAsExistsError("azurerm_thing", id.ID()) // want "this already-exists error is returned even when the provider is set to overwrite - put the check inside `if !<client>.Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources`"
	}

	return nil
}

func resourceGuardedCreate(meta any) error {
	id := thingId{}
	if !meta.(*clients.Client).Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
		if exists(id) {
			return tf.ImportAsExistsError("azurerm_thing", id.ID())
		}
	}

	return nil
}

// asking in the same condition is asking
func resourceCombinedCreate(meta any) error {
	id := thingId{}
	if exists(id) && !meta.(*clients.Client).Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
		return tf.ImportAsExistsError("azurerm_thing", id.ID())
	}

	return nil
}

// asking somewhere else in the create is not
func resourceElsewhereCreate(meta any) error {
	id := thingId{}
	if !meta.(*clients.Client).Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
		_ = exists(id)
	}

	if exists(id) {
		return tf.ImportAsExistsError("azurerm_thing", id.ID()) // want `this already-exists error is returned even when the provider is set to overwrite`
	}

	return nil
}

type ThingResource struct{}

func (r ThingResource) ResourceType() string { return "azurerm_thing" }

func (r ThingResource) Create() func(metadata sdk.ResourceMetaData) error {
	return func(metadata sdk.ResourceMetaData) error {
		id := thingId{}
		if exists(id) {
			return metadata.ResourceRequiresImport(r.ResourceType(), id) // want `this already-exists error is returned even when the provider is set to overwrite`
		}

		return nil
	}
}

func (r ThingResource) Update() func(metadata sdk.ResourceMetaData) error {
	return func(metadata sdk.ResourceMetaData) error {
		id := thingId{}
		if !metadata.Client.Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
			if exists(id) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}
		}

		return nil
	}
}
