package ingest_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/openshift/traust-sdk/go/v1/enums"
	"github.com/openshift/traust-sdk/go/v1/ingest"
	"github.com/openshift/traust-sdk/go/v1/ingest/ingesttest"
	"github.com/openshift/traust-sdk/go/v1/types"
)

// testFingerprint is a 64-hex fingerprint constant assembled at runtime so no
// key-shaped literal sits in the tree for forge secret scanners.
var testFingerprint = strings.Repeat("aa11bb22cc33dd44ee55ff66", 2) + "aa11bb22cc33dd44"

func TestConvertTriageReport(t *testing.T) {
	report := types.Triage{
		TriageCompleted: "2026-07-09",
		TriageContext: types.TriageContext{
			HarnessVersion: "0.27.0",
		},
		Findings: []types.TriageFinding{
			{
				Id:         "f001",
				OrigId:     "TEST_REPO-abc1234-001",
				Verdict:    enums.VerdictTruePositive,
				Rationale:  "reachable handler",
				FirstLinks: []string{"pkg/handler.go:73"},
			},
			{
				Id:        "f002",
				OrigId:    "TEST_REPO-abc1234-002",
				Verdict:   enums.VerdictUndetermined,
				Rationale: "unclear exploit path",
			},
			{
				Id:      "f003",
				Verdict: enums.VerdictDuplicate,
			},
		},
	}

	fpIndex := map[string]string{
		"TEST_REPO-abc1234-001": testFingerprint,
	}
	out := ingest.ConvertTriageReport(report, "findings/repo-a/triage.json", "2026-07-11T12:00:00+00:00", fpIndex)
	if len(out.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out.Events))
	}
	if len(out.NeedsReview) != 1 {
		t.Fatalf("expected 1 needs_review item, got %d", len(out.NeedsReview))
	}

	event := out.Events[0]
	if _, hasID := event["event_id"]; hasID {
		t.Fatal("events must not include event_id; server stamps it")
	}
	if event["finding_ref"] != "TEST_REPO-abc1234-001" {
		t.Fatalf("unexpected finding_ref: %v", event["finding_ref"])
	}
	if event["fingerprint"] != fpIndex["TEST_REPO-abc1234-001"] {
		t.Fatalf("expected fingerprint stamped from index, got %v", event["fingerprint"])
	}
	if _, hasAlgo := event["fingerprint_algo"]; hasAlgo {
		t.Fatal("SDK must not set fingerprint_algo; the ledger stamps it")
	}
	disposition, ok := event["disposition"].(map[string]interface{})
	if !ok || disposition["validity"] != "confirmed" {
		t.Fatalf("unexpected disposition: %v", event["disposition"])
	}

	review := out.NeedsReview[0]
	if review["queue_reason"] != "undetermined_finding" {
		t.Fatalf("unexpected queue_reason: %v", review["queue_reason"])
	}
}

func TestConvertValidationReport(t *testing.T) {
	title := "SQL injection"
	impact := "arbitrary file read"
	report := types.Validation{
		Metadata: types.ValidationMetadata{
			Date:           "2026-07-09",
			HarnessVersion: "0.28.0",
		},
		ValidatedFindings: []types.ValidatedFinding{
			{
				SourceId:       "pkg:src/FIND-001",
				Technique:      "path-traversal",
				Verdict:        enums.ValidationVerdictConfirmed,
				Title:          &title,
				ObservedImpact: &impact,
			},
			{
				SourceId:  "pkg:src/FIND-002",
				Technique: "replay",
				Verdict:   enums.ValidationVerdictRefuted,
			},
			{
				SourceId: "pkg:src/FIND-003",
				Verdict:  enums.ValidationVerdictInconclusive,
			},
		},
	}

	out := ingest.ConvertValidationReport(report, "validations/repo-a/validation.json", "2026-07-11T12:00:00+00:00", nil)
	if len(out.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out.Events))
	}
	for _, event := range out.Events {
		if _, hasID := event["event_id"]; hasID {
			t.Fatal("events must not include event_id; server stamps it")
		}
	}
}

func TestConvertVerificationReport(t *testing.T) {
	report := types.Verification{
		Metadata: types.VerificationMetadata{
			Date:           "2026-07-09",
			HarnessVersion: "0.28.0",
		},
		VerifiedFindings: []types.VerifiedFinding{
			{
				OriginalId: "FIND-001",
				Verdict:    "resolved",
				Evidence: types.VerifiedFindingEvidence{
					Explanation: "patch removes sink",
				},
			},
			{
				OriginalId: "FIND-002",
				Verdict:    "partially_resolved",
				CrossRepo: &types.CrossRepo{
					Propagation: "pending",
				},
				Evidence: types.VerifiedFindingEvidence{
					Explanation: "fix landed upstream, pending downstream",
				},
			},
			{
				OriginalId: "FIND-003",
				Verdict:    "false_positive",
				Evidence: types.VerifiedFindingEvidence{
					Explanation: "not reachable after refactor",
				},
			},
			{
				OriginalId: "FIND-004",
				Verdict:    "bogus",
			},
			{
				OriginalId: "",
				Verdict:    "resolved",
			},
		},
		Regressions: []types.Regression{
			{Id: "REG-001", Title: "new issue"},
		},
	}

	out := ingest.ConvertVerificationReport(report, "verifications/repo-a/verification.json", "2026-07-11T12:00:00+00:00", nil)
	if len(out.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(out.Events))
	}
	if len(out.Skipped) != 2 {
		t.Fatalf("expected 2 skipped findings (bogus + empty id), got %d", len(out.Skipped))
	}

	event := out.Events[0]
	source, ok := event["source"].(map[string]interface{})
	if !ok || source["type"] != "verification_report" {
		t.Fatalf("unexpected source: %v", event["source"])
	}
	if event["occurred_at"] != "2026-07-09T00:00:00+00:00" {
		t.Fatalf("unexpected occurred_at: %v", event["occurred_at"])
	}

	partial := out.Events[1]
	disposition, ok := partial["disposition"].(map[string]interface{})
	if !ok || disposition["resolution"] != "fix_in_progress" {
		t.Fatalf("expected fix_in_progress for pending cross-repo, got %v", partial["disposition"])
	}

	fp := out.Events[2]
	fpDisp, ok := fp["disposition"].(map[string]interface{})
	if !ok || fpDisp["validity"] != "false_positive" {
		t.Fatalf("expected false_positive validity, got %v", fp["disposition"])
	}
}

func TestBatchSubmit_RoundTrip(t *testing.T) {
	provider := ingesttest.NewStaticProvider().
		WithSubmitResponse(ingest.SubmitResponse{ID: "batch-001", Status: "accepted"})
	client := ingest.NewClient(provider)

	resp, err := client.BatchSubmit(context.Background(), "repo-a", ingest.BatchSubmitInput{
		SourceRef:  "findings/repo-a/triage.json",
		RecordedAt: "2026-01-16T00:00:00+00:00",
		Events: []map[string]interface{}{
			{"finding_ref": "FIND-001"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "batch-001" {
		t.Fatalf("expected ID batch-001, got %s", resp.ID)
	}

	calls := provider.SubmitCalls()
	if len(calls) != 1 || calls[0].LayerID != "repo-a" {
		t.Fatalf("unexpected submit calls: %+v", calls)
	}
}

func TestBatchSubmit_DecodesNewResponseFields(t *testing.T) {
	ec := 2
	provider := ingesttest.NewStaticProvider().
		WithSubmitResponse(ingest.SubmitResponse{
			ID:         "batch-002",
			Status:     "accepted",
			EventCount: &ec,
			EventIDs:   []string{"evt-001", "evt-002"},
			QueueAdded: 1,
		})
	client := ingest.NewClient(provider)

	resp, err := client.BatchSubmit(context.Background(), "repo-a", ingest.BatchSubmitInput{
		SourceRef:  "findings/repo-a/triage.json",
		RecordedAt: "2026-01-16T00:00:00+00:00",
		Events: []map[string]interface{}{
			{"finding_ref": "FIND-001"},
			{"finding_ref": "FIND-002"},
		},
		NeedsReview: []map[string]interface{}{
			{"finding_ref": "FIND-003", "queue_reason": "undetermined_finding"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.EventIDs) != 2 {
		t.Fatalf("expected 2 event_ids, got %d", len(resp.EventIDs))
	}
	if resp.EventIDs[0] != "evt-001" || resp.EventIDs[1] != "evt-002" {
		t.Fatalf("unexpected event_ids: %v", resp.EventIDs)
	}
	if resp.QueueAdded != 1 {
		t.Fatalf("expected queue_added 1, got %d", resp.QueueAdded)
	}
}

func TestSubmitTriageReport_BatchSubmitsConvertedEvents(t *testing.T) {
	provider := ingesttest.NewStaticProvider().
		WithSubmitResponse(ingest.SubmitResponse{ID: "triage-001", Status: "accepted"})
	client := ingest.NewClient(provider)

	in := fixtureTriageInput()
	converted := ingest.ConvertTriageReport(in.Report, in.SourceRef, in.RecordedAt, in.FingerprintIndex)

	_, err := client.BatchSubmit(context.Background(), in.LayerID, ingest.BatchSubmitInput{
		SourceRef:   in.SourceRef,
		RecordedAt:  in.RecordedAt,
		Events:      converted.Events,
		NeedsReview: converted.NeedsReview,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := provider.SubmitCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 submit call, got %d", len(calls))
	}
	if calls[0].LayerID != "repo-a" {
		t.Fatalf("expected layer repo-a, got %s", calls[0].LayerID)
	}

	var body ingest.BatchSubmitInput
	if err := json.Unmarshal(calls[0].Payload, &body); err != nil {
		t.Fatalf("decode batch body: %v", err)
	}
	if body.SourceRef != "findings/repo-a/triage.json" {
		t.Fatalf("unexpected source_ref: %s", body.SourceRef)
	}
	if len(body.Events) != 1 {
		t.Fatalf("expected 1 converted event, got %d", len(body.Events))
	}
}

func TestSubmitTriageReport_FingerprintCheck(t *testing.T) {
	provider := ingesttest.NewStaticProvider()
	client := ingest.NewClient(provider)

	// fixtureTriageInput carries no FingerprintIndex, so the derived event has
	// no fingerprint and the post-conversion event-level check must fail,
	// naming the offending finding_ref.
	_, err := client.SubmitTriageReport(context.Background(), fixtureTriageInput())
	if err == nil {
		t.Fatal("expected fingerprint error when no FingerprintIndex is supplied")
	}
	var ie *ingest.IngestError
	if !errors.As(err, &ie) || ie.Phase != ingest.PhaseFingerprint {
		t.Fatalf("expected fingerprint phase error, got %v", err)
	}
}

func TestSubmitTriageReport_FingerprintIndexStamped(t *testing.T) {
	provider := ingesttest.NewStaticProvider().
		WithSubmitResponse(ingest.SubmitResponse{ID: "triage-001", Status: "accepted"})
	client := ingest.NewClient(provider)

	in := fixtureTriageInput()
	// The triage finding's orig_id is the finding_ref the converter emits; the
	// caller supplies its scan-lane fingerprint via the index.
	in.FingerprintIndex = map[string]string{
		"ACME_API-abc1234-001": testFingerprint,
	}

	if _, err := client.SubmitTriageReport(context.Background(), in); err != nil {
		t.Fatalf("unexpected error with FingerprintIndex supplied: %v", err)
	}

	calls := provider.SubmitCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 submit call, got %d", len(calls))
	}
	var body ingest.BatchSubmitInput
	if err := json.Unmarshal(calls[0].Payload, &body); err != nil {
		t.Fatalf("decode batch body: %v", err)
	}
	if len(body.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(body.Events))
	}
	if body.Events[0]["fingerprint"] != in.FingerprintIndex["ACME_API-abc1234-001"] {
		t.Fatalf("event fingerprint not stamped from index: %v", body.Events[0]["fingerprint"])
	}
	if _, hasAlgo := body.Events[0]["fingerprint_algo"]; hasAlgo {
		t.Fatal("SDK must not set fingerprint_algo; the ledger stamps it")
	}
}

func TestSubmitTriageReport_SchemaValidation(t *testing.T) {
	provider := ingesttest.NewStaticProvider()
	client := ingest.NewClient(provider)

	_, err := client.SubmitTriageReport(context.Background(), ingest.TriageReportInput{
		ReportMeta: ingest.ReportMeta{
			LayerID:    "repo-a",
			SourceRef:  "findings/repo-a/triage.json",
			RecordedAt: "2026-01-16T00:00:00+00:00",
		},
	})
	if err == nil {
		t.Fatal("expected schema validation error for empty report")
	}
	var ie *ingest.IngestError
	if !errors.As(err, &ie) {
		t.Fatalf("expected *IngestError, got %T", err)
	}
	if ie.Phase != ingest.PhaseValidate {
		t.Fatalf("expected phase validate, got %s", ie.Phase)
	}
}

func TestResolveReviewItem_RoundTrip(t *testing.T) {
	path := "/v1/ledger/layers/repo-a/resolve"
	provider := ingesttest.NewStaticProvider().
		WithPostResponse(path, ingest.ResolveResponse{Resolved: true, Key: `["a","b","c","d"]`})
	client := ingest.NewClient(provider)

	resp, err := client.ResolveReviewItem(context.Background(), "repo-a", ingest.ResolveInput{
		Key:      `["a","b","c","d"]`,
		Decision: "confirmed",
		Note:     "reviewed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Resolved || resp.Key == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(provider.PostCalls()) != 1 || provider.PostCalls()[0].Path != path {
		t.Fatalf("unexpected post calls: %+v", provider.PostCalls())
	}
}

func TestSignLayer_RoundTrip(t *testing.T) {
	path := "/v1/ledger/layers/repo-a/sign"
	provider := ingesttest.NewStaticProvider().
		WithPostResponse(path, ingest.SignResponse{Status: "signed", Method: "cosign", LayerID: "repo-a"})
	client := ingest.NewClient(provider)

	resp, err := client.SignLayer(context.Background(), "repo-a", ingest.SignOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "signed" || resp.Method != "cosign" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(provider.PostCalls()) != 1 || provider.PostCalls()[0].Path != path {
		t.Fatalf("unexpected post calls: %+v", provider.PostCalls())
	}
}

func TestSignLayer_WithRekor(t *testing.T) {
	path := "/v1/ledger/layers/repo-a/sign?rekor=true"
	provider := ingesttest.NewStaticProvider().
		WithPostResponse(path, ingest.SignResponse{Status: "signed", Method: "cosign", LayerID: "repo-a"})
	client := ingest.NewClient(provider)

	resp, err := client.SignLayer(context.Background(), "repo-a", ingest.SignOpts{Rekor: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "signed" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(provider.PostCalls()) != 1 || provider.PostCalls()[0].Path != path {
		t.Fatalf("expected rekor path, got: %+v", provider.PostCalls())
	}
}

func TestComputeFingerprints_RoundTrip(t *testing.T) {
	provider := ingesttest.NewStaticProvider().
		WithPostResponse("/v1/ledger/fingerprint", ingest.FingerprintResponse{
			StampedCount: 1,
			Findings: []map[string]interface{}{
				{"id": "FIND-001", "fingerprint": "fp-abc"},
			},
		})
	client := ingest.NewClient(provider)

	resp, err := client.ComputeFingerprints(context.Background(), ingest.FingerprintInput{
		Findings: []map[string]interface{}{
			{"id": "FIND-001"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StampedCount != 1 {
		t.Fatalf("expected stamped_count 1, got %d", resp.StampedCount)
	}
}

func TestSubmitCountersign_RoundTrip(t *testing.T) {
	provider := ingesttest.NewStaticProvider().
		WithCountersignResponse(ingest.SubmitResponse{ID: "cs-001", Status: "accepted"})
	client := ingest.NewClient(provider)

	resp, err := client.SubmitCountersign(context.Background(), ingest.CountersignInput{
		EventMeta: ingest.EventMeta{
			LayerID:    "repo-a",
			RecordedAt: "2026-01-16T00:00:00+00:00",
		},
		FindingRef:    "FIND-001",
		Verdict:       "true_positive",
		Justification: "Reviewed source and confirmed exploit path.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "cs-001" {
		t.Fatalf("expected ID cs-001, got %s", resp.ID)
	}
	calls := provider.IngestCalls()
	last := calls[len(calls)-1]
	if last.Meta.Kind != ingest.KindCountersign {
		t.Fatalf("expected kind countersign, got %s", last.Meta.Kind)
	}
}

func TestBatchSubmit_DecodeError(t *testing.T) {
	provider := &badSubmitProvider{data: []byte(`not json`)}
	client := ingest.NewClient(provider)

	_, err := client.BatchSubmit(context.Background(), "repo-a", ingest.BatchSubmitInput{
		SourceRef:  "ref",
		RecordedAt: "2026-01-16T00:00:00+00:00",
	})
	if err == nil {
		t.Fatal("expected decode error")
	}
	var ie *ingest.IngestError
	if !errors.As(err, &ie) || ie.Phase != ingest.PhaseDecode {
		t.Fatalf("expected decode phase error, got %v", err)
	}
}

type badSubmitProvider struct {
	data []byte
}

func (p *badSubmitProvider) Ingest(context.Context, ingest.IngestMeta, []byte) ([]byte, error) {
	return nil, errors.New("not used")
}

func (p *badSubmitProvider) Submit(context.Context, string, []byte) ([]byte, error) {
	return p.data, nil
}

func (p *badSubmitProvider) Post(context.Context, string, []byte) ([]byte, error) {
	return nil, errors.New("not used")
}

func fixtureTriageInput() ingest.TriageReportInput {
	raw := `{
		"layer_id": "repo-a",
		"source_ref": "findings/repo-a/triage.json",
		"recorded_at": "2026-01-16T00:00:00+00:00",
		"report": {
			"triage_completed": "2026-01-16",
			"triage_context": {
				"environment": "development",
				"harness_version": "0.299.0",
				"repo": "https://github.com/acme/api",
				"votes_per_finding": 3
			},
			"summary": {
				"by_severity": {"critical": 0, "high": 1, "low": 0, "medium": 0},
				"duplicates": 0,
				"false_positives": 0,
				"hardening": 0,
				"input_count": 1,
				"true_positives": 1,
				"undetermined": 0
			},
			"findings": [{
				"id": "f001",
				"orig_id": "ACME_API-abc1234-001",
				"title": "SQL injection",
				"verdict": "true_positive",
				"rationale": "Parameterized query missing in login handler",
				"first_links": ["src/login.go:42"],
				"fingerprint": "fp-test-001"
			}]
		}
	}`
	var in ingest.TriageReportInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		panic(err)
	}
	return in
}

func TestConvertOccurredAtNeverDoublesTheTimeComponent(t *testing.T) {
	cases := []struct {
		name       string
		date       string
		recordedAt string
		want       string
	}{
		{"bare date padded", "2026-07-09", "2026-07-11T12:00:00+00:00", "2026-07-09T00:00:00+00:00"},
		{"timestamp passes through", "2026-07-09T14:30:00Z", "2026-07-11T12:00:00+00:00", "2026-07-09T14:30:00Z"},
		{"offset passes through", "2026-07-09T14:30:00+02:00", "", "2026-07-09T14:30:00+02:00"},
		{"empty falls back", "", "2026-07-11T12:00:00+00:00", "2026-07-11T12:00:00+00:00"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := types.Verification{
				Metadata: types.VerificationMetadata{Date: tc.date, HarnessVersion: "0.28.0"},
				VerifiedFindings: []types.VerifiedFinding{{
					OriginalId: "FIND-001",
					Verdict:    "resolved",
					Evidence:   types.VerifiedFindingEvidence{Explanation: "patch removes sink"},
				}},
			}

			out := ingest.ConvertVerificationReport(report, "verifications/v.json", tc.recordedAt, nil)
			if len(out.Events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(out.Events))
			}
			got, _ := out.Events[0]["occurred_at"].(string)
			if got != tc.want {
				t.Fatalf("occurred_at: got %q, want %q", got, tc.want)
			}
			if strings.Count(got, "T") > 1 {
				t.Fatalf("duplicate time component: %q", got)
			}
		})
	}
}
