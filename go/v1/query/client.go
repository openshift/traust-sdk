package query

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/openshift/traust-sdk/go/v1/types"
)

// Client binds a Provider for typed read methods.
type Client struct {
	p Provider
}

// NewClient creates a Client backed by the given Provider.
func NewClient(p Provider) *Client { return &Client{p: p} }

// Health checks ledger service availability.
func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	body, err := c.p.Query(ctx, http.MethodGet, "/healthz")
	if err != nil {
		return HealthResponse{}, fmt.Errorf("query: health: %w", err)
	}

	var out HealthResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return HealthResponse{}, fmt.Errorf("query: decode health response: %w", err)
	}
	return out, nil
}

// GetLayer returns the full layer document for layerID.
func (c *Client) GetLayer(ctx context.Context, layerID string) (types.Layer, error) {
	path := fmt.Sprintf("/v1/ledger/layers/%s", layerID)
	body, err := c.p.Query(ctx, http.MethodGet, path)
	if err != nil {
		return types.Layer{}, fmt.Errorf("query: get layer: %w", err)
	}

	var out types.Layer
	if err := json.Unmarshal(body, &out); err != nil {
		return types.Layer{}, fmt.Errorf("query: decode layer response: %w", err)
	}
	return out, nil
}

// GetFindings returns resolved finding dispositions for layerID.
func (c *Client) GetFindings(ctx context.Context, layerID string) (FindingsResponse, error) {
	path := fmt.Sprintf("/v1/ledger/layers/%s/findings", layerID)
	body, err := c.p.Query(ctx, http.MethodGet, path)
	if err != nil {
		return FindingsResponse{}, fmt.Errorf("query: get findings: %w", err)
	}

	var out FindingsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return FindingsResponse{}, fmt.Errorf("query: decode findings response: %w", err)
	}
	return out, nil
}

// ListFindings returns paginated findings across all layers.
func (c *Client) ListFindings(ctx context.Context, opts ListFindingsOpts) (BulkFindingsResponse, error) {
	path := "/v1/ledger/findings"
	values := url.Values{}
	if opts.Cursor != nil && *opts.Cursor != "" {
		values.Set("cursor", *opts.Cursor)
	}
	if opts.Limit > 0 {
		values.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.SinceEpoch != nil {
		values.Set("since_epoch", strconv.Itoa(*opts.SinceEpoch))
	}
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.p.Query(ctx, http.MethodGet, path)
	if err != nil {
		return BulkFindingsResponse{}, fmt.Errorf("query: list findings: %w", err)
	}

	var out BulkFindingsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return BulkFindingsResponse{}, fmt.Errorf("query: decode bulk findings response: %w", err)
	}
	return out, nil
}

// ListEvents returns the raw event log for layerID with optional filtering.
func (c *Client) ListEvents(ctx context.Context, layerID string, opts ListEventsOpts) (EventsResponse, error) {
	path := fmt.Sprintf("/v1/ledger/layers/%s/events", layerID)
	values := url.Values{}
	if opts.FindingRef != nil && *opts.FindingRef != "" {
		values.Set("finding_ref", *opts.FindingRef)
	}
	if opts.SourceType != nil && *opts.SourceType != "" {
		values.Set("source_type", *opts.SourceType)
	}
	if opts.Limit > 0 {
		values.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		values.Set("offset", strconv.Itoa(opts.Offset))
	}
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.p.Query(ctx, http.MethodGet, path)
	if err != nil {
		return EventsResponse{}, fmt.Errorf("query: list events: %w", err)
	}

	var out EventsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return EventsResponse{}, fmt.Errorf("query: decode events response: %w", err)
	}
	return out, nil
}

// VerifyLayer checks layer integrity and optional signatures.
func (c *Client) VerifyLayer(ctx context.Context, layerID string, opts VerifyOpts) (VerifyResponse, error) {
	path := fmt.Sprintf("/v1/ledger/layers/%s/verify", layerID)
	if opts.CheckSignatures {
		path += "?check_signatures=true"
	}

	body, err := c.p.Query(ctx, http.MethodGet, path)
	if err != nil {
		return VerifyResponse{}, fmt.Errorf("query: verify layer: %w", err)
	}

	var out VerifyResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return VerifyResponse{}, fmt.Errorf("query: decode verify response: %w", err)
	}
	return out, nil
}

// ListLayers returns all layer IDs known to the ledger service.
func (c *Client) ListLayers(ctx context.Context) (LayerListResponse, error) {
	body, err := c.p.Query(ctx, http.MethodGet, "/v1/ledger/layers")
	if err != nil {
		return LayerListResponse{}, fmt.Errorf("query: list layers: %w", err)
	}

	var out LayerListResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return LayerListResponse{}, fmt.Errorf("query: decode layer list response: %w", err)
	}
	return out, nil
}
