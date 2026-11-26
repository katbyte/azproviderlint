// Package acceptance is a minimal stub of the provider's internal/acceptance and its check package for the analyzer tests.
package acceptance

import "testing"

type TestStep struct {
	Config      string
	Check       any
	ImportState bool
}

type TestData struct{ ResourceName string }

func BuildTestData(t *testing.T, resourceType, label string) TestData { return TestData{} }

func (TestData) ResourceTest(t *testing.T, r any, steps []TestStep) {}

func (TestData) ResourceSequentialTest(t *testing.T, r any, steps []TestStep) {}

func (TestData) DataSourceTest(t *testing.T, steps []TestStep) {}

func (TestData) ImportStep(ignore ...string) TestStep { return TestStep{} }

func (TestData) ImportBlockWithIDStep(expectNonEmptyPlan bool) TestStep { return TestStep{} }

func (TestData) RequiresImportErrorStep(config func(TestData) string) TestStep { return TestStep{} }

func ComposeTestCheckFunc(checks ...any) any { return nil }

type That struct{}

func CheckThat(resourceName string) That { return That{} }

func (That) ExistsInAzure(r any) any { return nil }

func (That) Key(name string) That { return That{} }

func (That) HasValue(value string) any { return nil }
