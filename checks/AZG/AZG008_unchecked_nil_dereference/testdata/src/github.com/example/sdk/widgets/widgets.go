// Package widgets is a go-azure-sdk-shaped stand-in used by the AZG008 fixtures: one PUT
// method marshalling its input, a wrapper delegating to it, and a GET that sends no body.
package widgets

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-azure-sdk/sdk/client"
)

type Widget struct{ Name *string }

type Envelope struct{ Widget Widget }

type WidgetsClient struct{ Client client.Client }

func (c WidgetsClient) CreateOrUpdate(ctx context.Context, id string, input Widget) error {
	opts := client.RequestOptions{HttpMethod: http.MethodPut, Path: id}
	req, err := c.Client.NewRequest(ctx, opts)
	if err != nil {
		return err
	}
	return req.Marshal(input)
}

func (c WidgetsClient) ForRegionCreateOrUpdateThenPoll(ctx context.Context, id string, input Widget) error {
	return c.CreateOrUpdate(ctx, id, input)
}

func (c WidgetsClient) Get(ctx context.Context, id string, input Widget) error {
	opts := client.RequestOptions{HttpMethod: http.MethodGet, Path: id}
	_, err := c.Client.NewRequest(ctx, opts)
	return err
}

func (c WidgetsClient) Put(ctx context.Context, id string, input Envelope) error {
	opts := client.RequestOptions{HttpMethod: http.MethodPut, Path: id}
	req, err := c.Client.NewRequest(ctx, opts)
	if err != nil {
		return err
	}
	return req.Marshal(input)
}

func (c WidgetsClient) PutPtr(ctx context.Context, id string, input *Widget) error {
	opts := client.RequestOptions{HttpMethod: http.MethodPut, Path: id}
	req, err := c.Client.NewRequest(ctx, opts)
	if err != nil {
		return err
	}
	return req.Marshal(input)
}
