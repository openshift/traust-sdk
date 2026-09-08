package types_test

import (
	"encoding/json"
	"testing"

	"github.com/openshift/traust-sdk/go/v1/enums"
	"github.com/openshift/traust-sdk/go/v1/types"
	"github.com/openshift/traust-sdk/go/v1/validate"
)

func TestReportRoundTrip(t *testing.T) {
	original := minimalReport()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := validate.Validate("report", data); err != nil {
		t.Fatalf("validate report schema: %v", err)
	}

	var roundTripped types.Report
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if roundTripped.Title != original.Title {
		t.Errorf("title: got %q, want %q", roundTripped.Title, original.Title)
	}
	if roundTripped.Metadata.Date != original.Metadata.Date {
		t.Errorf("metadata.date: got %q, want %q", roundTripped.Metadata.Date, original.Metadata.Date)
	}
	if roundTripped.Metadata.Scope != original.Metadata.Scope {
		t.Errorf("metadata.scope: got %q, want %q", roundTripped.Metadata.Scope, original.Metadata.Scope)
	}
	if roundTripped.ExecutiveSummary.Prose != original.ExecutiveSummary.Prose {
		t.Errorf("executive_summary.prose: got %q, want %q", roundTripped.ExecutiveSummary.Prose, original.ExecutiveSummary.Prose)
	}
	if len(roundTripped.SeverityCriteria) != len(original.SeverityCriteria) {
		t.Errorf("severity_criteria length: got %d, want %d", len(roundTripped.SeverityCriteria), len(original.SeverityCriteria))
	}
	if len(roundTripped.FindingsSummary) != len(original.FindingsSummary) {
		t.Errorf("findings_summary length: got %d, want %d", len(roundTripped.FindingsSummary), len(original.FindingsSummary))
	}
	if len(roundTripped.RemediationRoadmap) != len(original.RemediationRoadmap) {
		t.Errorf("remediation_roadmap length: got %d, want %d", len(roundTripped.RemediationRoadmap), len(original.RemediationRoadmap))
	}
}

func TestLayerRoundTrip(t *testing.T) {
	original := minimalLayer()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := validate.Validate("layer", data); err != nil {
		t.Fatalf("validate layer schema: %v", err)
	}

	var roundTripped types.Layer
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if roundTripped.Metadata.AuditReport != original.Metadata.AuditReport {
		t.Errorf("metadata.audit_report: got %q, want %q", roundTripped.Metadata.AuditReport, original.Metadata.AuditReport)
	}
	if roundTripped.Metadata.Repository != original.Metadata.Repository {
		t.Errorf("metadata.repository: got %q, want %q", roundTripped.Metadata.Repository, original.Metadata.Repository)
	}
}

func TestTriageRoundTrip(t *testing.T) {
	original := minimalTriage()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := validate.Validate("triage", data); err != nil {
		t.Fatalf("validate triage schema: %v", err)
	}

	var roundTripped types.Triage
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if roundTripped.TriageCompleted != original.TriageCompleted {
		t.Errorf("triage_completed: got %q, want %q", roundTripped.TriageCompleted, original.TriageCompleted)
	}
	if roundTripped.TriageContext.Repo != original.TriageContext.Repo {
		t.Errorf("triage_context.repo: got %q, want %q", roundTripped.TriageContext.Repo, original.TriageContext.Repo)
	}
}

func minimalReport() types.Report {
	longDefinition := "Exploitable vulnerability with severe impact on confidentiality, integrity, or availability of the system."
	return types.Report{
		Title: "Security Audit Report",
		Metadata: types.ReportMetadata{
			Date:  "2026-08-12",
			Scope: "Full source code review of the application repository and its dependencies.",
		},
		ExecutiveSummary: types.ExecutiveSummary{
			Prose: "This security audit assessed the target application for common vulnerability classes. No critical issues were identified during this review cycle.",
			SeverityCounts: types.ExecutiveSummarySeverityCounts{
				Critical:      0,
				High:          0,
				Medium:        0,
				Low:           0,
				Informational: 0,
			},
		},
		SeverityCriteria: []types.SeverityCriterion{
			{Level: enums.SeverityCritical, Definition: longDefinition},
			{Level: enums.SeverityHigh, Definition: longDefinition},
			{Level: enums.SeverityMedium, Definition: longDefinition},
			{Level: enums.SeverityLow, Definition: longDefinition},
		},
		Findings: []types.ReportFinding{},
		FindingsSummary: []types.SeverityCountEntry{
			{Severity: enums.SeverityCritical, Count: 0, FindingIds: []string{}},
			{Severity: enums.SeverityHigh, Count: 0, FindingIds: []string{}},
			{Severity: enums.SeverityMedium, Count: 0, FindingIds: []string{}},
			{Severity: enums.SeverityLow, Count: 0, FindingIds: []string{}},
		},
		RemediationRoadmap: []types.RoadmapItem{
			{
				Priority:  "low",
				Action:    "Continue periodic security reviews and dependency updates.",
				Addresses: []string{"general-hygiene"},
			},
		},
	}
}

func minimalLayer() types.Layer {
	return types.Layer{
		Metadata: types.LayerMetadata{
			AuditReport:    "example-security-audit.json",
			Repository:     "https://github.com/example/repo",
			Created:        "2026-08-12T10:00:00Z",
			HarnessVersion: "0.2.0",
		},
		Events:      []types.Event{},
		NeedsReview: []types.ReviewItem{},
	}
}

func minimalTriage() types.Triage {
	return types.Triage{
		TriageCompleted: "2026-08-12",
		TriageContext: types.TriageContext{
			Environment:     "local development environment with full source access",
			VotesPerFinding: 1,
			Repo:            "https://github.com/example/repo",
			HarnessVersion:  "0.2.0",
		},
		Summary: types.TriageSummary{
			InputCount:     0,
			TruePositives:  0,
			Hardening:      0,
			FalsePositives: 0,
			Undetermined:   0,
			Duplicates:     0,
			BySeverity: types.TriageSummaryBySeverity{
				Critical: 0,
				High:     0,
				Medium:   0,
				Low:      0,
			},
		},
		Findings: []types.TriageFinding{},
	}
}
