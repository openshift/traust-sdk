// Package ingesttest provides test doubles for the ingest package.
package ingesttest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openshift/traust-sdk/go/v1/ingest"
)

// StaticProvider is a test double that returns pre-configured responses.
// It implements ingest.Provider.
type StaticProvider struct {
	ingestResponses map[ingest.IngestKind][]byte
	submitResponse  []byte
	postResponses   map[string][]byte
	err             error
	ingestCalls     []IngestCall
	submitCalls     []SubmitCall
	postCalls       []PostCall
}

// IngestCall records one invocation of Ingest.
type IngestCall struct {
	Meta    ingest.IngestMeta
	Payload []byte
}

// SubmitCall records one invocation of Submit.
type SubmitCall struct {
	LayerID string
	Payload []byte
}

// PostCall records one invocation of Post.
type PostCall struct {
	Path    string
	Payload []byte
}

// NewStaticProvider creates an empty StaticProvider.
func NewStaticProvider() *StaticProvider {
	return &StaticProvider{
		ingestResponses: make(map[ingest.IngestKind][]byte),
		postResponses:   make(map[string][]byte),
	}
}

// WithResponse sets the response for a specific human-lane kind.
func (p *StaticProvider) WithResponse(kind ingest.IngestKind, resp interface{}) *StaticProvider {
	data, _ := json.Marshal(resp)
	p.ingestResponses[kind] = data
	return p
}

// WithSubmitResponse sets the response for batch submit calls.
func (p *StaticProvider) WithSubmitResponse(resp ingest.SubmitResponse) *StaticProvider {
	data, _ := json.Marshal(resp)
	p.submitResponse = data
	return p
}

// WithTriageResponse is a shorthand for WithSubmitResponse.
func (p *StaticProvider) WithTriageResponse(resp ingest.SubmitResponse) *StaticProvider {
	return p.WithSubmitResponse(resp)
}

// WithCountersignResponse is a shorthand for WithResponse(KindCountersign, resp).
func (p *StaticProvider) WithCountersignResponse(resp ingest.SubmitResponse) *StaticProvider {
	return p.WithResponse(ingest.KindCountersign, resp)
}

// WithPostResponse sets the response for a specific Post path.
func (p *StaticProvider) WithPostResponse(path string, resp interface{}) *StaticProvider {
	data, _ := json.Marshal(resp)
	p.postResponses[path] = data
	return p
}

// WithSignResponse sets the response for POST /v1/ledger/layers/{layerID}/sign.
func (p *StaticProvider) WithSignResponse(layerID string, resp ingest.SignResponse) *StaticProvider {
	return p.WithPostResponse(fmt.Sprintf("/v1/ledger/layers/%s/sign", layerID), resp)
}

// WithError makes all Provider calls return the given error.
func (p *StaticProvider) WithError(err error) *StaticProvider {
	p.err = err
	return p
}

// Ingest implements ingest.Provider.
func (p *StaticProvider) Ingest(_ context.Context, meta ingest.IngestMeta, payload []byte) ([]byte, error) {
	p.ingestCalls = append(p.ingestCalls, IngestCall{Meta: meta, Payload: append([]byte(nil), payload...)})
	if p.err != nil {
		return nil, p.err
	}
	if resp, ok := p.ingestResponses[meta.Kind]; ok {
		return resp, nil
	}
	return json.Marshal(ingest.SubmitResponse{ID: "test-id", Status: "accepted"})
}

// Submit implements ingest.Provider.
func (p *StaticProvider) Submit(_ context.Context, layerID string, payload []byte) ([]byte, error) {
	p.submitCalls = append(p.submitCalls, SubmitCall{LayerID: layerID, Payload: append([]byte(nil), payload...)})
	if p.err != nil {
		return nil, p.err
	}
	if len(p.submitResponse) > 0 {
		return p.submitResponse, nil
	}
	return json.Marshal(ingest.SubmitResponse{ID: "batch-test-id", Status: "accepted"})
}

// Post implements ingest.Provider.
func (p *StaticProvider) Post(_ context.Context, path string, payload []byte) ([]byte, error) {
	p.postCalls = append(p.postCalls, PostCall{Path: path, Payload: append([]byte(nil), payload...)})
	if p.err != nil {
		return nil, p.err
	}
	if resp, ok := p.postResponses[path]; ok {
		return resp, nil
	}
	return []byte("{}"), nil
}

// IngestCalls returns all recorded Ingest invocations.
func (p *StaticProvider) IngestCalls() []IngestCall { return p.ingestCalls }

// SubmitCalls returns all recorded Submit invocations.
func (p *StaticProvider) SubmitCalls() []SubmitCall { return p.submitCalls }

// PostCalls returns all recorded Post invocations.
func (p *StaticProvider) PostCalls() []PostCall { return p.postCalls }

// Calls is an alias for IngestCalls for backward compatibility in tests.
func (p *StaticProvider) Calls() []IngestCall { return p.ingestCalls }

// FixtureTriageResponse returns a typical triage submission response.
func FixtureTriageResponse() ingest.SubmitResponse {
	ec := 1
	return ingest.SubmitResponse{
		ID:         "triage-001",
		Status:     "accepted",
		EventCount: &ec,
		EventIDs:   []string{"evt-triage-001"},
		QueueAdded: 0,
	}
}

// FixtureSignResponse returns a typical layer signing response.
func FixtureSignResponse() ingest.SignResponse {
	return ingest.SignResponse{Status: "signed", Method: "cosign", LayerID: "repo-a"}
}
