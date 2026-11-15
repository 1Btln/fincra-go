// Package httpclient provides the authenticated verb-oriented request
// interface used by resource services and SDK callers. The root SDK client
// keeps request execution and transport policy private.
package httpclient

import (
	"context"
	"net/http"

	"github.com/1Btln/fincra-go/internal/api"
)

// DoFunc executes a request using the SDK client's shared request pipeline.
type DoFunc func(ctx context.Context, method, path string, body, out any, opts ...func(*http.Request)) error

// Client adapts the SDK's low-level Do operation to standard HTTP verbs.
type Client struct {
	do         DoFunc
	businessID string
}

// New returns a verb-oriented client backed by do.
func New(do DoFunc, businessID string) *Client {
	return &Client{do: do, businessID: businessID}
}

// BusinessID returns the configured default business or sub-account ID.
func (c *Client) BusinessID() string { return c.businessID }

// Get sends a GET request without a body.
func (c *Client) Get(ctx context.Context, path string, out any, opts ...func(*http.Request)) error {
	return c.do(ctx, http.MethodGet, path, nil, out, opts...)
}

// Post sends a POST request.
func (c *Client) Post(ctx context.Context, path string, body, out any, opts ...func(*http.Request)) error {
	return c.do(ctx, http.MethodPost, path, body, out, opts...)
}

// Put sends a PUT request.
func (c *Client) Put(ctx context.Context, path string, body, out any, opts ...func(*http.Request)) error {
	return c.do(ctx, http.MethodPut, path, body, out, opts...)
}

// Patch sends a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body, out any, opts ...func(*http.Request)) error {
	return c.do(ctx, http.MethodPatch, path, body, out, opts...)
}

// Delete sends a DELETE request without a body.
func (c *Client) Delete(ctx context.Context, path string, out any, opts ...func(*http.Request)) error {
	return c.do(ctx, http.MethodDelete, path, nil, out, opts...)
}

var _ api.HTTPClient = (*Client)(nil)
