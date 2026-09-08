package query_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/openshift/traust-sdk/go/v1/query"
)

func TestHTTPClient_GetLayer(t *testing.T) {
	var gotPath, gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"events":[],"metadata":{"layer_id":"repo-a"},"needs_review":[]}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL, query.WithBearerToken("tok-123"))

	layer, err := client.GetLayer(context.Background(), "repo-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/ledger/layers/repo-a" {
		t.Fatalf("expected /v1/ledger/layers/repo-a, got %s", gotPath)
	}
	if gotAuth != "Bearer tok-123" {
		t.Fatalf("expected Bearer tok-123, got %s", gotAuth)
	}
	if len(layer.Events) != 0 {
		t.Fatalf("expected empty events, got %d", len(layer.Events))
	}
}

func TestHTTPClient_ListEvents(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"events":[],"total":0,"layer_id":"repo-a"}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)

	resp, err := client.ListEvents(context.Background(), "repo-a", query.ListEventsOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/ledger/layers/repo-a/events" {
		t.Fatalf("expected /v1/ledger/layers/repo-a/events, got %s", gotPath)
	}
	if resp.LayerID != "repo-a" {
		t.Fatalf("expected layer_id repo-a, got %s", resp.LayerID)
	}
}

func TestHTTPClient_ListEvents_WithParams(t *testing.T) {
	var gotPath, gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"events":[],"total":0,"layer_id":"repo-a"}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)
	findingRef := "FIND-001"
	sourceType := "triage_report"

	_, err := client.ListEvents(context.Background(), "repo-a", query.ListEventsOpts{
		FindingRef: &findingRef,
		SourceType: &sourceType,
		Limit:      25,
		Offset:     10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/ledger/layers/repo-a/events" {
		t.Fatalf("expected /v1/ledger/layers/repo-a/events, got %s", gotPath)
	}

	values, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if values.Get("finding_ref") != "FIND-001" {
		t.Fatalf("expected finding_ref=FIND-001, got %s", gotQuery)
	}
	if values.Get("source_type") != "triage_report" {
		t.Fatalf("expected source_type=triage_report, got %s", gotQuery)
	}
	if values.Get("limit") != "25" {
		t.Fatalf("expected limit=25, got %s", gotQuery)
	}
	if values.Get("offset") != "10" {
		t.Fatalf("expected offset=10, got %s", gotQuery)
	}
}

func TestHTTPClient_GetFindings(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"findings":[],"summary":{"by_validity":{},"by_resolution":{}},"ledger_only":true}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)

	_, err := client.GetFindings(context.Background(), "repo-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/ledger/layers/repo-a/findings" {
		t.Fatalf("expected /v1/ledger/layers/repo-a/findings, got %s", gotPath)
	}
}

func TestHTTPClient_ListFindings(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"layers":[],"total_findings":0,"has_more":false}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)

	_, err := client.ListFindings(context.Background(), query.ListFindingsOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/ledger/findings" {
		t.Fatalf("expected /v1/ledger/findings, got %s", gotPath)
	}
}

func TestHTTPClient_ListFindings_WithParams(t *testing.T) {
	var gotPath, gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"layers":[],"total_findings":0,"has_more":false}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)
	cursor := "repo-b"
	since := 42

	_, err := client.ListFindings(context.Background(), query.ListFindingsOpts{
		Cursor:     &cursor,
		Limit:      50,
		SinceEpoch: &since,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/ledger/findings" {
		t.Fatalf("expected /v1/ledger/findings, got %s", gotPath)
	}

	values, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if values.Get("cursor") != "repo-b" {
		t.Fatalf("expected cursor=repo-b, got %s", gotQuery)
	}
	if values.Get("limit") != "50" {
		t.Fatalf("expected limit=50, got %s", gotQuery)
	}
	if values.Get("since_epoch") != "42" {
		t.Fatalf("expected since_epoch=42, got %s", gotQuery)
	}
}

func TestHTTPClient_VerifyLayer(t *testing.T) {
	tests := []struct {
		name     string
		opts     query.VerifyOpts
		wantPath string
	}{
		{
			name:     "without signatures",
			opts:     query.VerifyOpts{},
			wantPath: "/v1/ledger/layers/repo-a/verify",
		},
		{
			name:     "with signatures",
			opts:     query.VerifyOpts{CheckSignatures: true},
			wantPath: "/v1/ledger/layers/repo-a/verify?check_signatures=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				if r.URL.RawQuery != "" {
					gotPath += "?" + r.URL.RawQuery
				}
				_, _ = w.Write([]byte(`{"passed":true,"findings":[],"checked_at":"2026-01-16T00:00:00+00:00"}`))
			}))
			defer srv.Close()

			client := query.NewHTTPClient(srv.URL)
			_, err := client.VerifyLayer(context.Background(), "repo-a", tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("expected %s, got %s", tt.wantPath, gotPath)
			}
		})
	}
}

func TestHTTPClient_Health(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)

	resp, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/healthz" {
		t.Fatalf("expected /healthz, got %s", gotPath)
	}
	if resp.Status != "healthy" {
		t.Fatalf("expected status healthy, got %s", resp.Status)
	}
}

func TestHTTPClient_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"layer not found"}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL)
	_, err := client.GetLayer(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}

	var se *query.StatusError
	if !errors.As(err, &se) {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}
	if se.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", se.StatusCode)
	}
}

func TestHTTPClient_AuthHeader(t *testing.T) {
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer srv.Close()

	client := query.NewHTTPClient(srv.URL, query.WithBearerToken("user-tok"))
	_, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer user-tok" {
		t.Fatalf("expected Bearer user-tok, got %s", gotAuth)
	}
}
