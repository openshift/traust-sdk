package ingest

import (
	"github.com/openshift/traust-sdk/go/v1/types"
)

// ReportMeta is shared metadata for machine-lane report submissions.
type ReportMeta struct {
	LayerID    string `json:"layer_id"`
	SourceRef  string `json:"source_ref"`
	RecordedAt string `json:"recorded_at"`
}

// TriageReportInput submits a triage report for event derivation.
type TriageReportInput struct {
	ReportMeta
	Report types.Triage `json:"report"`
	// FingerprintIndex maps a finding reference (the finding_ref emitted by the
	// converter) to its ledger-computed fingerprint. Triage reports do not carry
	// fingerprints on their findings — the fingerprint is a scan-lane identity —
	// so the caller supplies it here and the converter stamps it onto each event.
	// Not serialized: it never travels on the wire, only the derived events do.
	FingerprintIndex map[string]string `json:"-"`
}

// ValidationReportInput submits a validation report for event derivation.
type ValidationReportInput struct {
	ReportMeta
	Report types.Validation `json:"report"`
	// FingerprintIndex maps finding_ref to its ledger-computed fingerprint; see
	// TriageReportInput.FingerprintIndex. Not serialized.
	FingerprintIndex map[string]string `json:"-"`
}

// VerificationReportInput submits a verification report for event derivation.
type VerificationReportInput struct {
	ReportMeta
	Report types.Verification `json:"report"`
	// FingerprintIndex maps finding_ref to its ledger-computed fingerprint; see
	// TriageReportInput.FingerprintIndex. Not serialized.
	FingerprintIndex map[string]string `json:"-"`
}

// BatchSubmitInput is the body for POST /v1/ledger/layers/{layer_id}/submit.
type BatchSubmitInput struct {
	SourceRef   string                   `json:"source_ref"`
	RecordedAt  string                   `json:"recorded_at"`
	Events      []map[string]interface{} `json:"events"`
	NeedsReview []map[string]interface{} `json:"needs_review,omitempty"`
}

// ResolveInput is the body for POST /v1/ledger/layers/{layer_id}/resolve.
type ResolveInput struct {
	Key      string `json:"key"`
	Decision string `json:"decision"`
	Note     string `json:"note,omitempty"`
}

// FingerprintInput is the body for POST /v1/ledger/fingerprint.
type FingerprintInput struct {
	Findings   []map[string]interface{} `json:"findings"`
	Repository string                   `json:"repository,omitempty"`
}

// EventMeta is the shared prefix for human-lane event bodies.
type EventMeta struct {
	LayerID    string `json:"layer_id"`
	RecordedAt string `json:"recorded_at"`
}

// CountersignInput carries a human countersign decision for a finding.
// Actor identity is resolved server-side from the auth context.
type CountersignInput struct {
	EventMeta
	FindingRef         string `json:"finding_ref,omitempty"`
	FindingFingerprint string `json:"finding_fingerprint,omitempty"`
	Verdict            string `json:"verdict,omitempty"`
	Decision           string `json:"decision,omitempty"`
	Rationale          string `json:"rationale,omitempty"`
	Justification      string `json:"justification,omitempty"`
}

// SeverityInput carries a human severity override for a finding.
type SeverityInput struct {
	EventMeta
	FindingRef    string `json:"finding_ref"`
	Severity      string `json:"severity"`
	Rationale     string `json:"rationale,omitempty"`
	Justification string `json:"justification,omitempty"`
}

// SignOpts configures the SignLayer request.
type SignOpts struct {
	Rekor bool
}
