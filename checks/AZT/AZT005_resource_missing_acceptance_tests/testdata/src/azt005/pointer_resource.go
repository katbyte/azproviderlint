package azt005

type PointerResource struct{}

func (r *PointerResource) ResourceType() string {
	return "azurerm_pointer"
}

func (r *PointerResource) Update() {}
