package ingest

import (
	"context"
	"net/http"

	"github.com/openshift/traust-sdk/go/v1/ingest/internal/transport"
)

// StatusError is returned when the ledger service responds with a non-2xx
// HTTP status. Use errors.As to inspect the status code and response body.
type StatusError = transport.StatusError

// HTTPOption configures the built-in HTTP transport.
type HTTPOption func(*transport.HTTPProvider)

// WithHTTPClient sets a custom *http.Client (e.g. for timeouts, TLS config).
func WithHTTPClient(c *http.Client) HTTPOption {
	return HTTPOption(transport.WithHTTPClient(c))
}

// WithBearerToken sets a static bearer token for all requests.
func WithBearerToken(token string) HTTPOption {
	return HTTPOption(transport.WithBearerToken(token))
}

// WithAuthFunc sets a custom auth function called before each request.
// Use for token refresh, mutual TLS headers, or other dynamic auth.
func WithAuthFunc(fn func(*http.Request)) HTTPOption {
	return HTTPOption(transport.WithAuthFunc(fn))
}

// NewHTTPClient creates a Client backed by the ledger service's HTTP API.
//
//	client := ingest.NewHTTPClient("https://ledger.example.com",
//	    ingest.WithBearerToken(os.Getenv("LEDGER_TOKEN")),
//	)
//	resp, err := client.SubmitTriageReport(ctx, ingest.TriageReportInput{...})
func NewHTTPClient(baseURL string, opts ...HTTPOption) *Client {
	tOpts := make([]transport.Option, len(opts))
	for i, o := range opts {
		tOpts[i] = transport.Option(o)
	}
	return NewClient(&httpAdapter{t: transport.New(baseURL, tOpts...)})
}

// httpAdapter bridges the internal transport to the Provider interface.
type httpAdapter struct {
	t *transport.HTTPProvider
}

func (a *httpAdapter) Ingest(ctx context.Context, meta IngestMeta, payload []byte) ([]byte, error) {
	return a.t.Ingest(ctx, string(meta.Kind), meta.ContractsVersion, payload)
}

func (a *httpAdapter) Submit(ctx context.Context, layerID string, payload []byte) ([]byte, error) {
	return a.t.Submit(ctx, layerID, payload)
}

func (a *httpAdapter) Post(ctx context.Context, path string, payload []byte) ([]byte, error) {
	return a.t.Post(ctx, path, payload)
}
