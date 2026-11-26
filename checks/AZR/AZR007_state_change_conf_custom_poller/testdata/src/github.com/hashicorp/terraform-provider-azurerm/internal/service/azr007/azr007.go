package azr007

import (
	"context"
	"time"

	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

// Should be flagged: pointer StateChangeConf composite literal.
func invalidStateChangeConf(ctx context.Context) {
	stateConf := &pluginsdk.StateChangeConf{ // want `StateChangeConf`
		Pending: []string{"Creating"},
		Target:  []string{"Created"},
		Timeout: 10 * time.Minute,
	}
	_, _ = stateConf.WaitForStateContext(ctx)
}

// Should be flagged: value StateChangeConf composite literal.
func invalidStateChangeConfNoPointer(ctx context.Context) {
	stateConf := pluginsdk.StateChangeConf{ // want `StateChangeConf`
		Pending: []string{"Deleting"},
		Target:  []string{"Deleted"},
		Timeout: 5 * time.Minute,
	}
	_, _ = stateConf.WaitForStateContext(ctx)
}

// Should be flagged: empty StateChangeConf composite literal.
func invalidEmptyStateChangeConf(ctx context.Context) {
	conf := &pluginsdk.StateChangeConf{} // want `StateChangeConf`
	_, _ = conf.WaitForStateContext(ctx)
}

// Should be flagged: direct retry.StateChangeConf usage (the underlying SDK type).
func invalidRetryStateChangeConf(ctx context.Context) {
	stateConf := &retry.StateChangeConf{ // want `StateChangeConf`
		Pending: []string{"Updating"},
		Target:  []string{"Updated"},
		Timeout: 5 * time.Minute,
	}
	_, _ = stateConf.WaitForStateContext(ctx)
}

// Should be flagged: aliased import of resource.StateChangeConf (aliases the SDK type).
func invalidResourceStateChangeConf(ctx context.Context) {
	stateConf := sdk.StateChangeConf{ // want `StateChangeConf`
		Pending: []string{"Pending"},
		Target:  []string{"Done"},
		Timeout: 5 * time.Minute,
	}
	_, _ = stateConf.WaitForStateContext(ctx)
}

// Should NOT be flagged: a custom poller with a different type.
type MyCustomPoller struct{}

func (p *MyCustomPoller) PollUntilDone(ctx context.Context) error {
	return nil
}

func validCustomPoller() {
	p := &MyCustomPoller{}
	_ = p.PollUntilDone(context.Background())
}

// StateChangeConf shares the SDK type's name but lives in this package.
type StateChangeConf struct{ Target []string }

// Should NOT be flagged: a same-named type from a different package.
func validLocalStateChangeConf() {
	conf := &StateChangeConf{Target: []string{"Done"}}
	_ = conf
}

// Should NOT be flagged: an unnamed composite literal.
func validAnonymousLiteral() {
	conf := struct{ Target []string }{Target: []string{"Done"}}
	_ = conf
}

// Should be flagged: the provider's wrapper, and the SDK helpers it wraps.
func retries(ctx context.Context) error {
	if err := pluginsdk.Retry(time.Minute, func() *pluginsdk.RetryError { // want `prefer a Custom Poller over Retry`
		return pluginsdk.RetryableError(nil)
	}); err != nil {
		return err
	}

	_ = retry.Retry(time.Minute, nil) // want `prefer a Custom Poller over Retry`

	return retry.RetryContext(ctx, time.Minute, nil) // want `prefer a Custom Poller over RetryContext`
}

type client struct{}

func (client) Retry(time.Duration) error { return nil }

func Retry() error { return nil }

// Should NOT be flagged: a method or a function of another package that shares the name.
func validOtherRetries() {
	_ = client{}.Retry(time.Minute)
	_ = Retry()
}
