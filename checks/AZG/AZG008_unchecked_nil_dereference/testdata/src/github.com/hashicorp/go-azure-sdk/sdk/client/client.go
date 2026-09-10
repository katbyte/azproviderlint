// Package client is a minimal stand-in for go-azure-sdk's request builder: Marshal really
// serialises its argument, so the fact is derived rather than assumed.
package client

import (
	"context"
	"encoding/json"
)

type RequestOptions struct {
	ContentType         string
	ExpectedStatusCodes []int
	HttpMethod          string
	Path                string
}

type Request struct{ body []byte }

func (r *Request) Marshal(v interface{}) error {
	b, err := json.Marshal(v)
	r.body = b
	return err
}

func (*Request) Execute(ctx context.Context) (*Response, error) { return nil, nil }

type Response struct{}

type Client struct{}

func (Client) NewRequest(ctx context.Context, opts RequestOptions) (*Request, error) {
	return &Request{}, nil
}
