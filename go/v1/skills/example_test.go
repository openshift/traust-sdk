package skills_test

import (
	"context"
	"errors"
	"testing"

	"github.com/openshift/traust-sdk/go/v1/skills"
	"github.com/openshift/traust-sdk/go/v1/skills/skillstest"
)

func TestScan_Run(t *testing.T) {
	provider := skillstest.NewStaticProvider().
		WithScanResult(skillstest.FixtureReport())

	report, err := skills.Scan.Run(context.Background(), provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Title == "" {
		t.Fatal("expected non-empty report title")
	}
}

func TestClient_Scan(t *testing.T) {
	provider := skillstest.NewStaticProvider().
		WithScanResult(skillstest.FixtureReport())
	client := skills.NewClient(provider)

	report, err := client.Scan(context.Background(), skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Title != "Fixture Security Assessment" {
		t.Fatalf("got title %q, want %q", report.Title, "Fixture Security Assessment")
	}
}

func TestClient_Triage(t *testing.T) {
	provider := skillstest.NewStaticProvider().
		WithTriageResult(skillstest.FixtureTriage())
	client := skills.NewClient(provider)

	triage, err := client.Triage(context.Background(), skills.TriageInput{
		Repo: "https://github.com/org/repo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if triage.TriageContext.HarnessVersion != "0.24.0" {
		t.Fatalf("got harness version %q, want %q", triage.TriageContext.HarnessVersion, "0.24.0")
	}
}

func TestSkillError_UnregisteredSkill(t *testing.T) {
	provider := skillstest.NewStaticProvider()

	_, err := skills.Scan.Run(context.Background(), provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err == nil {
		t.Fatal("expected error for unregistered skill")
	}

	var se *skills.SkillError
	if !errors.As(err, &se) {
		t.Fatalf("expected *skills.SkillError, got %T", err)
	}
	if se.Phase != skills.PhaseExecute {
		t.Fatalf("expected phase %q, got %q", skills.PhaseExecute, se.Phase)
	}
	if se.Skill != skills.SkillSecureCodeAudit {
		t.Fatalf("expected skill %q, got %q", skills.SkillSecureCodeAudit, se.Skill)
	}
}
