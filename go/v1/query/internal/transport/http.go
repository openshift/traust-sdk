// Package transport implements the HTTP transport for ledger read endpoints.
package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StatusError is returned when the ledger service responds with a non-2xx status.
type StatusError struct {
	StatusCode int
	Body       []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("query: http status %d: %s", e.StatusCode, string(e.Body))
}

// HTTPProvider implements the ledger service HTTP read transport.
type HTTPProvider struct {
	baseURL string
	client  *http.Client
	auth    func(*http.Request)
}

// Option configures the HTTP provider.
type Option func(*HTTPProvider)

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(p *HTTPProvider) { p.client = c }
}

// WithBearerToken sets a static bearer token for all requests.
func WithBearerToken(token string) Option {
	return func(p *HTTPProvider) {
		p.auth = func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+token)
		}
	}
}

// WithAuthFunc sets a custom auth function called before each request.
func WithAuthFunc(fn func(*http.Request)) Option {
	return func(p *HTTPProvider) { p.auth = fn }
}

// New creates an HTTPProvider targeting the given ledger service base URL.
func New(baseURL string, opts ...Option) *HTTPProvider {
	p := &HTTPProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  http.DefaultClient,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Query issues an HTTP request to the ledger service read API.
func (p *HTTPProvider) Query(ctx context.Context, method, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("query: build request: %w", err)
	}

	if p.auth != nil {
		p.auth(req)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("query: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
	}

	return respBody, nil
}
