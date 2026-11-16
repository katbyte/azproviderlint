package azr011

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/azr011/stubs"
)

type MissingResource struct{ stubs.Resource } // want `MissingResource has no "var _ sdk\.Resource = MissingResource\{\}" line to have the compiler check it where it is declared`

// an update is optional and found when the provider runs, so the line names the interface that includes it
type MissingUpdatableResource struct { // want `MissingUpdatableResource has no "var _ sdk\.ResourceWithUpdate = MissingUpdatableResource\{\}" line`
	stubs.ResourceWithUpdate
}

type MissingDataSource struct{ stubs.DataSource } // want `MissingDataSource has no "var _ sdk\.DataSource = MissingDataSource\{\}" line`

type MissingFrameworkResource struct{ stubs.FrameworkResource } // want `MissingFrameworkResource has no "var _ sdk\.FrameworkWrappedResourceWithUpdate = MissingFrameworkResource\{\}" line`

// methods on the pointer make the pointer the thing to assert
type MissingAction struct{} // want `MissingAction has no "var _ sdk\.Action = &MissingAction\{\}" line`

func (*MissingAction) Invoke() {}

// declared in a group, the line goes after the group
type (
	MissingList struct{} // want `MissingList has no "var _ sdk\.FrameworkListWrappedResource = MissingList\{\}" line`

	listOptions struct{}
)

func (MissingList) List(sdk.ResourceMetadata) {}

type MissingEphemeral struct{} // want `MissingEphemeral has no "var _ sdk\.EphemeralResource = MissingEphemeral\{\}" line`

func (MissingEphemeral) Open() {}

// an interface from anywhere else does not check it against the sdk
type OtherInterfaceResource struct{ stubs.Resource } // want `OtherInterfaceResource has no "var _ sdk\.Resource = OtherInterfaceResource\{\}" line`

var _ interface{ ResourceType() string } = OtherInterfaceResource{}
