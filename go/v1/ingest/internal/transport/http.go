// Package transport implements the HTTP transport for the ledger service.
//
// Wire contract:
//
//	POST /v1/ledger/events                        — human lane (countersign, severity)
//	POST /v1/ledger/layers/{layer_id}/submit      — machine lane batch submit
//	POST /v1/ledger/layers/{layer_id}/resolve     — resolve needs_review item
//	POST /v1/ledger/fingerprint                   — stamp finding fingerprints
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const eventsPath = "/v1/ledger/events"

type eventEnvelope struct {
	Kind             string          `json:"kind"`
	ContractsVersion string          `json:"contracts_version"`
	Event            json.RawMessage `json:"event"`
}

// StatusError is returned when the ledger service responds with a non-2xx status.
type StatusError struct {
	StatusCode int
	Body       []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("ingest: http status %d: %s", e.StatusCode, string(e.Body))
}

// HTTPProvider implements the ledger service HTTP transport.
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

// Ingest wraps the payload in the human-lane event envelope and POSTs to
// POST /v1/ledger/events.
func (p *HTTPProvider) Ingest(ctx context.Context, kind, contractsVersion string, payload []byte) ([]byte, error) {
	body, err := json.Marshal(eventEnvelope{
		Kind:             kind,
		ContractsVersion: contractsVersion,
		Event:            payload,
	})
	if err != nil {
		return nil, fmt.Errorf("ingest: marshal envelope: %w", err)
	}
	return p.post(ctx, eventsPath, body)
}

// Submit POSTs a batch-submit body to POST /v1/ledger/layers/{layer_id}/submit.
func (p *HTTPProvider) Submit(ctx context.Context, layerID string, payload []byte) ([]byte, error) {
	path := fmt.Sprintf("/v1/ledger/layers/%s/submit", layerID)
	return p.post(ctx, path, payload)
}

// Post sends a raw JSON body to an absolute API path (e.g. /v1/ledger/fingerprint).
func (p *HTTPProvider) Post(ctx context.Context, path string, payload []byte) ([]byte, error) {
	return p.post(ctx, path, payload)
}

func (p *HTTPProvider) post(ctx context.Context, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ingest: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.auth != nil {
		p.auth(req)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ingest: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ingest: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &StatusError{
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
	}

	return respBody, nil
}
