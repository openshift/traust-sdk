package validate

import (
	"slices"
	"testing"
)

func minimalValidReport() []byte {
	return []byte(`{
  "title": "Widget Security Assessment",
  "metadata": {
    "date": "2026-01-01",
    "scope": "unit test scope"
  },
  "executive_summary": {
    "prose": "This executive summary describes the security assessment outcome in sufficient detail.",
    "severity_counts": {
      "critical": 0,
      "high": 0,
      "medium": 0,
      "low": 0,
      "informational": 0
    }
  },
  "severity_criteria": [
    {"level": "critical", "definition": "Critical severity definition text."},
    {"level": "high", "definition": "High severity definition text here."},
    {"level": "medium", "definition": "Medium severity definition text."},
    {"level": "low", "definition": "Low severity definition text here."}
  ],
  "findings": [],
  "findings_summary": [
    {"severity": "critical", "count": 0, "finding_ids": []},
    {"severity": "high", "count": 0, "finding_ids": []},
    {"severity": "medium", "count": 0, "finding_ids": []},
    {"severity": "low", "count": 0, "finding_ids": []}
  ],
  "remediation_roadmap": [
    {
      "priority": "P1",
      "action": "Remediate identified issues promptly.",
      "addresses": ["none"]
    }
  ]
}`)
}

func minimalValidLayer() []byte {
	return []byte(`{
  "metadata": {
    "audit_report": "test-widget-security-audit.json",
    "repository": "https://example.com/repo",
    "created": "2026-06-01T00:00:00+00:00",
    "harness_version": "0.48.0"
  },
  "events": [],
  "needs_review": []
}`)
}

func TestValidateValidReport(t *testing.T) {
	if err := Validate("report", minimalValidReport()); err != nil {
		t.Fatalf("Validate(report): %v", err)
	}
}

func TestValidateInvalidReport(t *testing.T) {
	invalid := []byte(`{
  "title": "Widget Security Assessment",
  "metadata": {
    "scope": "unit test scope"
  }
}`)
	if err := Validate("report", invalid); err == nil {
		t.Fatal("expected validation error for report missing required fields")
	}
}

func TestValidateUnknownSchema(t *testing.T) {
	if err := Validate("not-a-schema", []byte(`{}`)); err == nil {
		t.Fatal("expected error for unknown schema")
	}
}

func TestSchemaNames(t *testing.T) {
	names := SchemaNames()
	for _, want := range []string{"report", "layer", "triage"} {
		if !slices.Contains(names, want) {
			t.Fatalf("SchemaNames() missing %q; got %v", want, names)
		}
	}
}

func TestCompiledSchemaValidation(t *testing.T) {
	compiled := MustCompile("layer")

	if err := compiled.Validate(minimalValidLayer()); err != nil {
		t.Fatalf("valid layer: %v", err)
	}

	invalid := []byte(`{"events": [], "needs_review": []}`)
	if err := compiled.Validate(invalid); err == nil {
		t.Fatal("expected validation error for layer missing metadata")
	}
}
