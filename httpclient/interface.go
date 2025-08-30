package httpclient

import (
	"context"
)

type HTTPClient interface {
	Get(ctx context.Context, url string, options ...RequestOption) (*Response, error)
	Post(ctx context.Context, url string, body interface{}, options ...RequestOption) (*Response, error)
	Put(ctx context.Context, url string, body interface{}, options ...RequestOption) (*Response, error)
	Delete(ctx context.Context, url string, options ...RequestOption) (*Response, error)
	Patch(ctx context.Context, url string, body interface{}, options ...RequestOption) (*Response, error)
}
