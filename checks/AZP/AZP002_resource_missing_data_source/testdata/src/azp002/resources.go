package azp002

// typed resources and data sources — correlated via ResourceType()

type CoveredTypedResource struct{}

func (r CoveredTypedResource) ResourceType() string {
	return "azurerm_covered_typed"
}

type CoveredTypedDataSource struct{}

func (r CoveredTypedDataSource) ResourceType() string {
	return "azurerm_covered_typed"
}

// CrossCoveredDataSource covers an untyped resource of the same name, exercising cross-flavour
// matching.
type CrossCoveredDataSource struct{}

func (r CrossCoveredDataSource) ResourceType() string {
	return "azurerm_cross_covered"
}

// MissingTypedResource declares its type name in a package-level var (the pattern some
// resources use so the name is shareable with locks helpers).
var missingTypedResourceName = "azurerm_missing_typed"

type MissingTypedResource struct{}

func (r MissingTypedResource) ResourceType() string {
	return missingTypedResourceName
}

// RunCommandResource is an invoke-style resource; its name matches the action-style suffix
// list, so it is never reported despite having no data source.
type RunCommandResource struct{}

func (r RunCommandResource) ResourceType() string {
	return "azurerm_thing_run_command"
}

// MissingAutoResource is registered via the generated auto-registration delegation and has no
// data source.
type MissingAutoResource struct{}

func (r MissingAutoResource) ResourceType() string {
	return "azurerm_missing_auto"
}

// CoveredAutoResource is registered directly and covered by the auto-registered data source
// below, exercising auto-registration entries counting towards the data source side.
type CoveredAutoResource struct{}

func (r CoveredAutoResource) ResourceType() string {
	return "azurerm_covered_auto"
}

type CoveredAutoDataSource struct{}

func (r CoveredAutoDataSource) ResourceType() string {
	return "azurerm_covered_auto"
}

type MissingAppendedResource struct{}

func (r MissingAppendedResource) ResourceType() string {
	return "azurerm_missing_appended"
}

// framework resources and data sources — same ResourceType() correlation

type CoveredFrameworkResource struct{}

func (r CoveredFrameworkResource) ResourceType() string {
	return "azurerm_covered_framework"
}

type CoveredFrameworkDataSource struct{}

func (r CoveredFrameworkDataSource) ResourceType() string {
	return "azurerm_covered_framework"
}

type MissingFrameworkResource struct{}

func (r MissingFrameworkResource) ResourceType() string {
	return "azurerm_missing_framework"
}
