package azr011alias

import (
	typed "github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
)

// the line is written with the name the file imports the sdk under
type MissingAction struct{} // want `MissingAction has no "var _ typed\.Action = MissingAction\{\}" line`

func (MissingAction) Invoke() {}

var unrelated typed.ResourceFunc
