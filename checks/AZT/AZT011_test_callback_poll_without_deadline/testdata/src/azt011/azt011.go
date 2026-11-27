package azt011

import (
	"context"
	"time"

	"azt011/acceptance"
)

type ThingResource struct{}

func steps(data acceptance.TestData, r ThingResource) {
	_ = data.CheckWithClientForResource(func(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
		return client.DeleteThenPoll(ctx, state.ID) // want `this callback's context has no deadline, and DeleteThenPoll refuses to poll without one - wrap it with context.WithTimeout first`
	}, data.ResourceName)

	_ = data.CheckWithClient(func(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
		return acceptance.Poller{}.PollUntilDone(ctx) // want `this callback's context has no deadline, and PollUntilDone refuses to poll without one`
	})

	// a callback that is a function of the package is read where it is written, once
	_ = data.CheckWithClient(r.disappears)
	_ = data.CheckWithClientWithoutResource(r.disappears)
	_ = data.CheckWithClient(bounded)

	// swapped for one with a deadline first
	_ = data.CheckWithClient(func(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		return client.DeleteThenPoll(ctx, state.ID)
	})

	// or polled with another that has one
	_ = data.CheckWithClient(func(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
		ctx2, cancel := context.WithDeadline(ctx, time.Now().Add(time.Hour))
		defer cancel()

		return client.DeleteThenPoll(ctx2, state.ID)
	})

	// a deadline given after the poll is too late
	_ = data.CheckWithClient(func(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
		if err := client.DeleteThenPoll(ctx, state.ID); err != nil { // want `this callback's context has no deadline`
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		return client.Get(ctx, state.ID)
	})

	// a call that does not poll needs none
	_ = data.CheckWithClient(func(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
		return client.Get(ctx, state.ID)
	})
}

func (ThingResource) disappears(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
	return client.DeleteThenPoll(ctx, state.ID) // want `this callback's context has no deadline`
}

func bounded(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Hour)
	defer cancel()

	return client.DeleteThenPoll(ctx, state.ID)
}

// a function no check is handed is given a context by whoever calls it
func (ThingResource) Destroy(ctx context.Context, client *acceptance.Client, state *acceptance.InstanceState) error {
	return client.DeleteThenPoll(ctx, state.ID)
}
