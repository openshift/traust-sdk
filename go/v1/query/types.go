package query

import "github.com/openshift/traust-sdk/go/v1/types"

// HealthResponse is returned by GET /healthz.
type HealthResponse struct {
	Status string `json:"status"`
}

// VerifyFinding describes one verification failure.
type VerifyFinding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// VerifyResponse is returned by GET /v1/ledger/layers/{layer_id}/verify.
type VerifyResponse struct {
	Passed    bool            `json:"passed"`
	Findings  []VerifyFinding `json:"findings"`
	CheckedAt string          `json:"checked_at"`
}

// FindingDisposition pairs a finding reference with its resolved disposition.
type FindingDisposition struct {
	FindingRef  string                  `json:"finding_ref"`
	Disposition types.ReportDisposition `json:"disposition"`
	EventCount  int                     `json:"event_count"`
	Fingerprint *string                 `json:"fingerprint,omitempty"`
	Orphan      *bool                   `json:"orphan,omitempty"`
}

// FindingsSummary summarizes disposition counts for a layer or bulk response.
// Validity and resolution counts reuse the generated contract types.
type FindingsSummary struct {
	ByValidity   types.DispositionSummaryByValidity   `json:"by_validity"`
	ByResolution types.DispositionSummaryByResolution `json:"by_resolution"`
}

// FindingsResponse is returned by GET /v1/ledger/layers/{layer_id}/findings.
type FindingsResponse struct {
	Findings   []FindingDisposition `json:"findings"`
	Summary    FindingsSummary      `json:"summary"`
	LedgerOnly bool                 `json:"ledger_only"`
}

// LayerFindings groups findings for one layer in a bulk response.
type LayerFindings struct {
	LayerID     string               `json:"layer_id"`
	MerkleRoot  *string              `json:"merkle_root,omitempty"`
	MerkleEpoch *int                 `json:"merkle_epoch,omitempty"`
	Findings    []FindingDisposition `json:"findings"`
	Summary     FindingsSummary      `json:"summary"`
	LedgerOnly  bool                 `json:"ledger_only"`
}

// BulkFindingsResponse is returned by GET /v1/ledger/findings.
type BulkFindingsResponse struct {
	Layers        []LayerFindings `json:"layers"`
	TotalFindings int             `json:"total_findings"`
	NextCursor    *string         `json:"next_cursor,omitempty"`
	HasMore       bool            `json:"has_more"`
}

// ListFindingsOpts configures paginated bulk findings queries.
type ListFindingsOpts struct {
	Cursor     *string
	Limit      int
	SinceEpoch *int
}

// VerifyOpts configures layer verification queries.
type VerifyOpts struct {
	CheckSignatures bool
}

// EventsResponse is returned by GET /v1/ledger/layers/{layer_id}/events.
type EventsResponse struct {
	Events  []types.Event `json:"events"`
	Total   int           `json:"total"`
	LayerID string        `json:"layer_id"`
}

// ListEventsOpts configures event history queries.
type ListEventsOpts struct {
	FindingRef *string
	SourceType *string
	Limit      int
	Offset     int
}

// LayerListResponse is returned by GET /v1/ledger/layers.
type LayerListResponse struct {
	Layers []string `json:"layers"`
}
