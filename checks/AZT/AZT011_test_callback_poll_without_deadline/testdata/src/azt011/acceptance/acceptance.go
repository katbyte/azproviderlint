// Package acceptance is a minimal stub of the provider's internal/acceptance for the analyzer tests.
package acceptance

import "context"

type Client struct{}

func (*Client) DeleteThenPoll(ctx context.Context, id string) error { return nil }

func (*Client) Get(ctx context.Context, id string) error { return nil }

type Poller struct{}

func (Poller) PollUntilDone(ctx context.Context) error { return nil }

type InstanceState struct{ ID string }

type ClientCheckFunc func(ctx context.Context, client *Client, state *InstanceState) error

type TestData struct{ ResourceName string }

func (TestData) CheckWithClient(check ClientCheckFunc) any { return nil }

func (TestData) CheckWithClientForResource(check ClientCheckFunc, resourceName string) any {
	return nil
}

func (TestData) CheckWithClientWithoutResource(check ClientCheckFunc) any { return nil }
