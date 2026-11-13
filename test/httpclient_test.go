package fincra_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/1Btln/fincra-go/internal/httpclient"
)

func TestVerbMethodsDelegateToDo(t *testing.T) {
	type requestCall struct {
		method string
		path   string
		body   any
		out    any
		header string
	}
	var calls []requestCall
	client := httpclient.New(func(_ context.Context, method, path string, body, out any, opts ...func(*http.Request)) error {
		req := &http.Request{Header: make(http.Header)}
		for _, opt := range opts {
			opt(req)
		}
		calls = append(calls, requestCall{
			method: method,
			path:   path,
			body:   body,
			out:    out,
			header: req.Header.Get("x-test-option"),
		})
		return nil
	}, "biz_test")

	ctx := context.Background()
	body := map[string]string{"value": "sample"}
	out := new(struct{})
	option := func(req *http.Request) { req.Header.Set("x-test-option", "forwarded") }
	requests := []struct {
		call   func() error
		method string
		path   string
		body   any
	}{
		{call: func() error { return client.Get(ctx, "/get", out, option) }, method: http.MethodGet, path: "/get"},
		{call: func() error { return client.Post(ctx, "/post", body, out) }, method: http.MethodPost, path: "/post", body: body},
		{call: func() error { return client.Put(ctx, "/put", body, out) }, method: http.MethodPut, path: "/put", body: body},
		{call: func() error { return client.Patch(ctx, "/patch", body, out) }, method: http.MethodPatch, path: "/patch", body: body},
		{call: func() error { return client.Delete(ctx, "/delete", out) }, method: http.MethodDelete, path: "/delete"},
	}

	for _, request := range requests {
		if err := request.call(); err != nil {
			t.Fatalf("%s %s: %v", request.method, request.path, err)
		}
	}
	if len(calls) != len(requests) {
		t.Fatalf("Do calls = %d, want %d", len(calls), len(requests))
	}
	for i, want := range requests {
		got := calls[i]
		if got.method != want.method || got.path != want.path || !reflect.DeepEqual(got.body, want.body) || got.out != out {
			t.Errorf("call[%d] = %+v, want method=%s path=%s body=%v out=%p", i, got, want.method, want.path, want.body, out)
		}
	}
	if calls[0].header != "forwarded" {
		t.Errorf("request option header = %q, want forwarded", calls[0].header)
	}
	if got := client.BusinessID(); got != "biz_test" {
		t.Errorf("BusinessID() = %q, want biz_test", got)
	}
}
