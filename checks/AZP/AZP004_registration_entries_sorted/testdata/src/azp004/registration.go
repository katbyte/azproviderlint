package azp004

import (
	"github.com/example/provider/framework"
	"github.com/example/provider/pluginsdk"
	"github.com/example/provider/sdk"
	"github.com/example/provider/typed"
)

type Registration struct{}

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_availability_set":    nil,
		"azurerm_dedicated_host":      nil,
		"azurerm_disk_encryption_set": nil,
		"azurerm_managed_disk":        nil,
		"azurerm_virtual_machine":     nil,
	}
}

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{}
}

func (r Registration) InvalidSupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_availability_set":    nil,
		"azurerm_dedicated_host":      nil,
		"azurerm_managed_disk":        nil,
		"azurerm_disk_encryption_set": nil, // want "registration entries should be sorted alphabetically: `azurerm_disk_encryption_set` should come before `azurerm_managed_disk`"
		"azurerm_ssh_public_key":      nil,
	}
}

func (r Registration) SupportedResourcesViaVariable() map[string]*pluginsdk.Resource {
	lookup := map[string]string{
		"z": "last",
		"a": "first",
	}
	_ = lookup

	resources := map[string]*pluginsdk.Resource{
		"azurerm_availability_set":    nil,
		"azurerm_dedicated_host":      nil,
		"azurerm_managed_disk":        nil,
		"azurerm_disk_encryption_set": nil, // want `registration entries should be sorted alphabetically`
		"azurerm_ssh_public_key":      nil,
	}

	return resources
}

func (r Registration) SectionedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		// CDN
		"azurerm_cdn_profile": nil,

		// FrontDoor
		"azurerm_cdn_frontdoor_custom_domain": nil,
		"azurerm_cdn_frontdoor_endpoint":      nil,
		"azurerm_cdn_frontdoor_profile":       nil,
	}
}

func (r Registration) InvalidSectionedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		// CDN
		"azurerm_cdn_profile": nil,

		// FrontDoor
		"azurerm_cdn_frontdoor_profile":       nil,
		"azurerm_cdn_frontdoor_custom_domain": nil, // want `registration entries should be sorted alphabetically`
		"azurerm_cdn_frontdoor_endpoint":      nil,
	}
}

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		ApiManagementResource{},
		WorkspaceResource{},
	}
}

func (r Registration) InvalidResources() []sdk.Resource {
	return []sdk.Resource{
		WorkspaceResource{},
		ApiManagementResource{}, // want "registration entries should be sorted alphabetically: `ApiManagementResource` should come before `WorkspaceResource`"
	}
}

func (r Registration) ResourcesViaVariable() []sdk.Resource {
	resources := []sdk.Resource{
		WorkspaceResource{},
		ApiManagementResource{}, // want `registration entries should be sorted alphabetically`
	}

	return resources
}

func (r Registration) InvalidPointerResources() []sdk.Resource {
	return []sdk.Resource{
		&WorkspaceResource{},
		&ApiManagementResource{}, // want `registration entries should be sorted alphabetically`
	}
}

func (r Registration) QualifiedResources() []sdk.Resource {
	return []sdk.Resource{
		typed.ComputeResource{},
		typed.NetworkResource{},
	}
}

func (r Registration) InvalidQualifiedResources() []sdk.Resource {
	return []sdk.Resource{
		typed.NetworkResource{},
		typed.ComputeResource{}, // want `registration entries should be sorted alphabetically`
	}
}

func (r Registration) FrameworkResources() []func() framework.Resource {
	return []func() framework.Resource{
		newApiManagementResource,
		newWorkspaceResource,
	}
}

func (r Registration) InvalidFrameworkResources() []func() framework.Resource {
	return []func() framework.Resource{
		newWorkspaceResource,
		newApiManagementResource, // want `registration entries should be sorted alphabetically`
	}
}

func (r Registration) WebsiteCategories() []string {
	return []string{
		"Compute",
		"Network",
	}
}

func (r Registration) InvalidWebsiteCategories() []string {
	return []string{
		"Network",
		"Compute", // want `registration entries should be sorted alphabetically`
	}
}

// CaseInsensitiveCategories is not reported: entries are ordered case-insensitively even though
// their casing differs.
func (r Registration) CaseInsensitiveCategories() []string {
	return []string{
		"apple",
		"Banana",
		"cherry",
	}
}

func (r Registration) InvalidCaseInsensitiveCategories() []string {
	return []string{
		"Cherry",
		"apple", // want `registration entries should be sorted alphabetically`
	}
}

func (r Registration) AttachedEntryComment() []string {
	return []string{
		"cherry",
		// zebra is special
		"zebra",
		"apple", // want `registration entries should be sorted alphabetically`
	}
}

// FirstEntryAttachedComment: a comment directly under the opening brace is attached to the first
// entry, so the suggested fix moves them together.
func (r Registration) FirstEntryAttachedComment() []string {
	return []string{
		// banana note
		"banana",
		"apple", // want `registration entries should be sorted alphabetically`
	}
}

// BlankLineCommentStartsSection: a comment with a blank line before it starts a new section even
// when a blank line also follows it, so these single-entry sections are each sorted and not flagged.
func (r Registration) BlankLineCommentStartsSection() []string {
	return []string{
		"zebra",

		// heading

		"apple",
	}
}

// UnresolvableEntriesNotReported has an entry whose sort key cannot be resolved, so the whole
// section is skipped rather than judged on the resolvable entries alone.
func (r Registration) UnresolvableEntriesNotReported() []sdk.Resource {
	return []sdk.Resource{
		WorkspaceResource{},
		makeResource("x"),
		ApiManagementResource{},
	}
}

func (r Registration) AppendedResourcesViaVariable() []sdk.Resource {
	resources := []sdk.Resource{
		WorkspaceResource{},
		ApiManagementResource{}, // want `registration entries should be sorted alphabetically`
	}

	resources = append(resources, buildResources()...)

	return resources
}

// BraceSharingNotFixed reports the unsorted entries but offers no auto-fix because the opening and
// closing braces share entry lines and a whole-line rewrite would corrupt the source.
func (r Registration) BraceSharingNotFixed() []sdk.Resource {
	return []sdk.Resource{WorkspaceResource{},
		ApiManagementResource{}} // want `registration entries should be sorted alphabetically`
}

// PartiallyFixableSections leaves the same-line section unchanged and fixes the safe section;
// each unsorted section gets its own report.
func (r Registration) PartiallyFixableSections() []string {
	return []string{
		"zebra", "apple", // want `registration entries should be sorted alphabetically`

		"dog",
		"cat", // want `registration entries should be sorted alphabetically`
	}
}

// SpanningBlockCommentNotFixed cannot be safely reordered by whole source lines.
func (r Registration) SpanningBlockCommentNotFixed() []string {
	return []string{
		"zebra", /* note
		spanning */"apple", // want `registration entries should be sorted alphabetically`
	}
}

type ApiManagementResource struct{}
type WorkspaceResource struct{}

func (ApiManagementResource) ResourceType() string { return "azurerm_api_management" }
func (WorkspaceResource) ResourceType() string     { return "azurerm_workspace" }

func newApiManagementResource() framework.Resource { return nil }
func newWorkspaceResource() framework.Resource     { return nil }

func makeResource(string) sdk.Resource { return nil }

func buildResources() []sdk.Resource { return nil }
