package ingest_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openshift/traust-sdk/go/v1/ingest"
)

func TestHTTPClient_BatchSubmitLane(t *testing.T) {
	var gotPath string
	var body ingest.BatchSubmitInput

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &body)
		w.Write([]byte(`{"id":"batch-ok","status":"accepted"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL, ingest.WithBearerToken("tok-123"))

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

	if gotPath != "/v1/ledger/layers/repo-a/submit" {
		t.Fatalf("expected batch submit path, got %s", gotPath)
	}
	if body.SourceRef != "findings/repo-a/triage.json" {
		t.Fatalf("expected source_ref in body, got %s", body.SourceRef)
	}
	if resp.ID != "batch-ok" {
		t.Fatalf("expected ID batch-ok, got %s", resp.ID)
	}
}

func TestHTTPClient_EventLane(t *testing.T) {
	var gotPath, gotAuth string
	var envelope struct {
		Kind             string          `json:"kind"`
		ContractsVersion string          `json:"contracts_version"`
		Event            json.RawMessage `json:"event"`
	}
	var event struct {
		LayerID    string `json:"layer_id"`
		FindingRef string `json:"finding_ref"`
		Verdict    string `json:"verdict"`
		RecordedAt string `json:"recorded_at"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &envelope)
		json.Unmarshal(envelope.Event, &event)
		w.Write([]byte(`{"id":"cs-1","status":"accepted"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL, ingest.WithBearerToken("user-tok"))

	_, err := client.SubmitCountersign(context.Background(), ingest.CountersignInput{
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

	if gotPath != "/v1/ledger/events" {
		t.Fatalf("expected /v1/ledger/events, got %s", gotPath)
	}
	if envelope.Kind != "countersign" {
		t.Fatalf("expected kind countersign in envelope, got %s", envelope.Kind)
	}
	if event.LayerID != "repo-a" {
		t.Fatalf("expected layer_id repo-a in event, got %s", event.LayerID)
	}
	if gotAuth != "Bearer user-tok" {
		t.Fatalf("expected Bearer user-tok, got %s", gotAuth)
	}
}

func TestHTTPClient_AllEventKindsRoute(t *testing.T) {
	var gotPath string
	var envelope struct {
		Kind string `json:"kind"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &envelope)
		w.Write([]byte(`{"id":"x","status":"accepted"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL)
	meta := ingest.EventMeta{LayerID: "repo-a", RecordedAt: "2026-01-16T00:00:00+00:00"}

	tests := []struct {
		kind string
		call func() error
	}{
		{"countersign", func() error {
			_, err := client.SubmitCountersign(context.Background(), ingest.CountersignInput{
				EventMeta: meta, FindingRef: "FIND-001", Verdict: "true_positive",
				Justification: "Reviewed source and confirmed exploit path.",
			})
			return err
		}},
		{"severity", func() error {
			_, err := client.SubmitSeverity(context.Background(), ingest.SeverityInput{
				EventMeta: meta, FindingRef: "FIND-001", Severity: "high",
				Rationale: "Reviewed source and confirmed exploit path.",
			})
			return err
		}},
	}

	for _, tt := range tests {
		if err := tt.call(); err != nil {
			t.Fatalf("kind %s: unexpected error: %v", tt.kind, err)
		}
		if gotPath != "/v1/ledger/events" {
			t.Fatalf("%s: expected /v1/ledger/events, got %s", tt.kind, gotPath)
		}
		if envelope.Kind != tt.kind {
			t.Fatalf("expected kind %s in envelope, got %s", tt.kind, envelope.Kind)
		}
	}
}

func TestHTTPClient_ResolveAndFingerprint(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/v1/ledger/layers/repo-a/resolve":
			w.Write([]byte(`{"resolved":true,"key":"k"}`))
		case "/v1/ledger/fingerprint":
			w.Write([]byte(`{"findings":[],"stamped_count":0}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL)

	if _, err := client.ResolveReviewItem(context.Background(), "repo-a", ingest.ResolveInput{
		Key: "k", Decision: "confirmed",
	}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if _, err := client.ComputeFingerprints(context.Background(), ingest.FingerprintInput{
		Findings: []map[string]interface{}{{"id": "FIND-001"}},
	}); err != nil {
		t.Fatalf("fingerprint: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %v", paths)
	}
}

func TestHTTPClient_SignLayer(t *testing.T) {
	var gotPath, gotAuth, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte(`{"status":"signed","method":"cosign","layer_id":"repo-a"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL, ingest.WithBearerToken("sign-tok"))

	resp, err := client.SignLayer(context.Background(), "repo-a", ingest.SignOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v1/ledger/layers/repo-a/sign" {
		t.Fatalf("expected sign path, got %s", gotPath)
	}
	if gotMethod != "POST" {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotAuth != "Bearer sign-tok" {
		t.Fatalf("expected Bearer sign-tok, got %s", gotAuth)
	}
	if resp.Status != "signed" || resp.LayerID != "repo-a" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHTTPClient_SignLayerRekor(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Write([]byte(`{"status":"signed","method":"cosign","layer_id":"repo-a"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL)
	_, err := client.SignLayer(context.Background(), "repo-a", ingest.SignOpts{Rekor: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "rekor=true" {
		t.Fatalf("expected rekor=true query, got %q", gotQuery)
	}
}

func TestHTTPClient_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"error":"validation failed"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL)
	_, err := client.BatchSubmit(context.Background(), "repo-a", ingest.BatchSubmitInput{
		SourceRef: "ref", RecordedAt: "2026-01-16T00:00:00+00:00",
	})
	if err == nil {
		t.Fatal("expected error for 422")
	}

	var se *ingest.StatusError
	if !errors.As(err, &se) {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}
	if se.StatusCode != 422 {
		t.Fatalf("expected status 422, got %d", se.StatusCode)
	}
}

func TestHTTPClient_KindInEnvelopeNotHeader(t *testing.T) {
	var gotKindHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKindHeader = r.Header.Get("X-Ingest-Kind")
		w.Write([]byte(`{"id":"x","status":"ok"}`))
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(srv.URL)
	client.SubmitCountersign(context.Background(), ingest.CountersignInput{
		EventMeta:  ingest.EventMeta{LayerID: "repo-a", RecordedAt: "2026-01-16T00:00:00+00:00"},
		FindingRef: "FIND-001", Verdict: "true_positive",
	})

	if gotKindHeader != "" {
		t.Fatal("kind should be in body envelope, not in headers")
	}
}
