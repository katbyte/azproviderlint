// Package clients is a minimal stub of the provider's internal/clients for the analyzer tests.
package clients

type UserFeatures struct {
	SkipImportCheckOnCreateAndAllowOverwritingExistingResources bool
}

type Client struct {
	Features UserFeatures
}
