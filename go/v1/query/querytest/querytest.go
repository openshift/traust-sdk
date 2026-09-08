// Package querytest provides test doubles for the query package.
package querytest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/traust-sdk/go/v1/enums"
	"github.com/openshift/traust-sdk/go/v1/query"
	"github.com/openshift/traust-sdk/go/v1/types"
)

// StaticProvider is a test double that returns pre-configured responses for
// ledger read paths. It implements query.Provider.
type StaticProvider struct {
	responses map[string][]byte
	err       error
	calls     []Call
}

// Call records one invocation of Query.
type Call struct {
	Method string
	Path   string
}

// NewStaticProvider creates an empty StaticProvider.
func NewStaticProvider() *StaticProvider {
	return &StaticProvider{responses: make(map[string][]byte)}
}

func (p *StaticProvider) withResponse(path string, resp interface{}) *StaticProvider {
	data, _ := json.Marshal(resp)
	p.responses[path] = data
	return p
}

// WithHealthResponse sets the response for GET /healthz.
func (p *StaticProvider) WithHealthResponse(resp query.HealthResponse) *StaticProvider {
	return p.withResponse("/healthz", resp)
}

// WithLayerResponse sets the response for GET /v1/ledger/layers/{layerID}.
func (p *StaticProvider) WithLayerResponse(layerID string, resp types.Layer) *StaticProvider {
	return p.withResponse(fmt.Sprintf("/v1/ledger/layers/%s", layerID), resp)
}

// WithFindingsResponse sets the response for GET /v1/ledger/layers/{layerID}/findings.
func (p *StaticProvider) WithFindingsResponse(layerID string, resp query.FindingsResponse) *StaticProvider {
	return p.withResponse(fmt.Sprintf("/v1/ledger/layers/%s/findings", layerID), resp)
}

// WithBulkFindingsResponse sets the response for GET /v1/ledger/findings.
func (p *StaticProvider) WithBulkFindingsResponse(resp query.BulkFindingsResponse) *StaticProvider {
	return p.withResponse("/v1/ledger/findings", resp)
}

// WithVerifyResponse sets the response for GET /v1/ledger/layers/{layerID}/verify.
func (p *StaticProvider) WithVerifyResponse(layerID string, resp query.VerifyResponse) *StaticProvider {
	return p.withResponse(fmt.Sprintf("/v1/ledger/layers/%s/verify", layerID), resp)
}

// WithEventsResponse sets the response for GET /v1/ledger/layers/{layerID}/events.
func (p *StaticProvider) WithEventsResponse(layerID string, resp query.EventsResponse) *StaticProvider {
	return p.withResponse(fmt.Sprintf("/v1/ledger/layers/%s/events", layerID), resp)
}

// WithError makes all Query calls return the given error.
func (p *StaticProvider) WithError(err error) *StaticProvider {
	p.err = err
	return p
}

// Query implements query.Provider.
func (p *StaticProvider) Query(_ context.Context, method, path string) ([]byte, error) {
	p.calls = append(p.calls, Call{Method: method, Path: path})
	if p.err != nil {
		return nil, p.err
	}
	if resp, ok := p.responses[path]; ok {
		return resp, nil
	}
	if i := strings.Index(path, "?"); i >= 0 {
		if resp, ok := p.responses[path[:i]]; ok {
			return resp, nil
		}
	}
	return json.Marshal(query.HealthResponse{Status: "healthy"})
}

// Calls returns all recorded invocations.
func (p *StaticProvider) Calls() []Call { return p.calls }

// FixtureHealthResponse returns a typical health response.
func FixtureHealthResponse() query.HealthResponse {
	return query.HealthResponse{Status: "healthy"}
}

// FixtureLayer returns a minimal layer document.
func FixtureLayer() types.Layer {
	return types.Layer{
		Events:      []types.Event{},
		Metadata:    types.LayerMetadata{Repository: "repo-a"},
		NeedsReview: []types.ReviewItem{},
	}
}

// FixtureFindingsResponse returns a typical layer findings response.
func FixtureFindingsResponse() query.FindingsResponse {
	return query.FindingsResponse{
		Findings: []query.FindingDisposition{
			{
				FindingRef: "FIND-001",
				Disposition: types.ReportDisposition{
					Events:      []string{"evt-1"},
					LastUpdated: "2026-01-16T00:00:00+00:00",
					Resolution:  enums.DispositionResolutionOpen,
					Validity:    enums.ValidityConfirmed,
				},
				EventCount: 1,
			},
		},
		Summary: query.FindingsSummary{
			ByValidity:   types.DispositionSummaryByValidity{Confirmed: 1},
			ByResolution: types.DispositionSummaryByResolution{Open: 1},
		},
		LedgerOnly: true,
	}
}

// FixtureBulkFindingsResponse returns a typical bulk findings page.
func FixtureBulkFindingsResponse() query.BulkFindingsResponse {
	return query.BulkFindingsResponse{
		Layers: []query.LayerFindings{
			{
				LayerID:    "repo-a",
				Findings:   FixtureFindingsResponse().Findings,
				Summary:    FixtureFindingsResponse().Summary,
				LedgerOnly: true,
			},
		},
		TotalFindings: 1,
		HasMore:       false,
	}
}

// FixtureVerifyResponse returns a typical verify response.
func FixtureVerifyResponse() query.VerifyResponse {
	return query.VerifyResponse{
		Passed:    true,
		Findings:  []query.VerifyFinding{},
		CheckedAt: "2026-01-16T00:00:00+00:00",
	}
}

// FixtureEventsResponse returns a typical layer events response.
func FixtureEventsResponse(layerID string) query.EventsResponse {
	return query.EventsResponse{
		Events: []types.Event{
			{
				EventId:    "evt-1",
				FindingRef: "FIND-001",
				Rationale:  "initial report",
				RecordedAt: "2026-01-16T00:00:00+00:00",
				Disposition: types.Disposition{
					Validity:   ptr(enums.EventValidityConfirmed),
					Resolution: ptr(enums.DispositionResolutionOpen),
				},
				Source: types.EventSource{
					Ref:  "report-1",
					Type: enums.SourceTypeTriageReport,
					Actor: types.Actor{
						Kind: enums.ActorKindHuman,
					},
				},
			},
		},
		Total:   1,
		LayerID: layerID,
	}
}

func ptr[T any](v T) *T { return &v }
