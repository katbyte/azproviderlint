package writebody

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-sdk/sdk/client"
)

type Model struct{ Name *string }

type Client struct{ Client client.Client }

func (c Client) CreateOrUpdate(ctx context.Context, id string, input Model) error { // want CreateOrUpdate:"body\\(2\\)"
	opts := client.RequestOptions{
		ContentType: "application/json; charset=utf-8",
		HttpMethod:  http.MethodPut,
		Path:        id,
	}
	req, err := c.Client.NewRequest(ctx, opts)
	if err != nil {
		return err
	}
	if err = req.Marshal(input); err != nil {
		return err
	}
	_, err = req.Execute(ctx)
	return err
}

func (c Client) Patch(ctx context.Context, id string, input Model) error { // want Patch:"body\\(2\\)"
	opts := client.RequestOptions{HttpMethod: http.MethodPatch, Path: id}
	req, _ := c.Client.NewRequest(ctx, opts)
	return req.Marshal(&input)
}

func (c Client) Action(ctx context.Context, id string, input Model) error { // want Action:"body\\(2\\)"
	opts := client.RequestOptions{HttpMethod: http.MethodPost, Path: id}
	req, _ := c.Client.NewRequest(ctx, opts)
	return req.Marshal(input)
}

// serialises but does not write: a GET that (oddly) marshals is not a body
func (c Client) Get(ctx context.Context, id string, input Model) error { // want Get:"serialises\\(2\\)"
	opts := client.RequestOptions{HttpMethod: http.MethodGet, Path: id}
	req, _ := c.Client.NewRequest(ctx, opts)
	return req.Marshal(input)
}

// writes but serialises nothing: a PUT with no body parameter
func (c Client) Touch(ctx context.Context, id string) error { // want Touch:"writes"
	opts := client.RequestOptions{HttpMethod: http.MethodPut, Path: id}
	_, err := c.Client.NewRequest(ctx, opts)
	return err
}

// delegation, one and two levels deep, whatever the wrapper is called
func (c Client) ForRegionThenPoll(ctx context.Context, id string, input Model) error { // want ForRegionThenPoll:"body\\(2\\)"
	return c.CallbackThenPoll(ctx, id, input, nil)
}

func (c Client) CallbackThenPoll(ctx context.Context, id string, input Model, cb func()) error { // want CallbackThenPoll:"body\\(2\\)"
	return c.CreateOrUpdate(ctx, id, input)
}

func update(c Client, m *Model) error { // want update:"body\\(1\\)"
	return c.Patch(context.Background(), "id", *m)
}

// writes (it calls a writer) but the parameter itself is not what is sent
func rename(c Client, id string, name string) error { // want rename:"writes"
	return c.CreateOrUpdate(context.Background(), id, Model{Name: &name})
}

// autorest style: method from a string literal, marshalling inside a closure
func (c Client) LegacyPut(ctx context.Context, parameters Model) error { // want LegacyPut:"body\\(1\\)"
	_ = autorest.CreatePreparer(autorest.AsPut(), autorest.WithJSON(parameters))
	return nil
}

func (c Client) LegacyGet(ctx context.Context, parameters Model) error { // want LegacyGet:"serialises\\(1\\)"
	_ = autorest.CreatePreparer(autorest.AsGet(), autorest.WithJSON(parameters))
	return nil
}

// a direct stdlib marshal of a parameter counts as serialising it
func encode(v Model) []byte { // want encode:"serialises\\(0\\)"
	b, _ := json.Marshal(v)
	return b
}

// no fact at all: neither writes nor serialises
func plain(v Model) string { return *v.Name }
