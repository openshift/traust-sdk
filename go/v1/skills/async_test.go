package skills_test

import (
	"context"
	"errors"
	"testing"

	"github.com/openshift/traust-sdk/go/v1/skills"
	"github.com/openshift/traust-sdk/go/v1/skills/skillstest"
)

func TestScan_DispatchCollect(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport())

	ctx := context.Background()
	ref, err := skills.Scan.Dispatch(ctx, provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ref.ID == "" {
		t.Fatal("expected non-empty job ref ID")
	}

	report, err := skills.Scan.Collect(ctx, provider, ref)
	if err != nil {
		t.Fatal(err)
	}
	if report.Title != "Fixture Security Assessment" {
		t.Fatalf("got title %q, want %q", report.Title, "Fixture Security Assessment")
	}
}

func TestTriage_DispatchCollect(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithTriageResult(skillstest.FixtureTriage())

	ctx := context.Background()
	ref, err := skills.Triage.Dispatch(ctx, provider, skills.TriageInput{
		Repo: "https://github.com/org/repo",
	})
	if err != nil {
		t.Fatal(err)
	}

	triage, err := skills.Triage.Collect(ctx, provider, ref)
	if err != nil {
		t.Fatal(err)
	}
	if triage.TriageContext.HarnessVersion != "0.24.0" {
		t.Fatalf("got harness version %q, want %q", triage.TriageContext.HarnessVersion, "0.24.0")
	}
}

func TestAsyncClient_ScanRoundTrip(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport())
	ac := skills.NewAsyncClient(provider)

	ctx := context.Background()
	ref, err := ac.DispatchScan(ctx, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}

	report, err := ac.CollectScan(ctx, ref)
	if err != nil {
		t.Fatal(err)
	}
	if report.Title != "Fixture Security Assessment" {
		t.Fatalf("got title %q, want %q", report.Title, "Fixture Security Assessment")
	}
}

func TestAsyncClient_TriageRoundTrip(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithTriageResult(skillstest.FixtureTriage())
	ac := skills.NewAsyncClient(provider)

	ctx := context.Background()
	ref, err := ac.DispatchTriage(ctx, skills.TriageInput{
		Repo: "https://github.com/org/repo",
	})
	if err != nil {
		t.Fatal(err)
	}

	triage, err := ac.CollectTriage(ctx, ref)
	if err != nil {
		t.Fatal(err)
	}
	if triage.TriageCompleted != "2026-01-01" {
		t.Fatalf("got triage completed %q, want %q", triage.TriageCompleted, "2026-01-01")
	}
}

func TestDispatch_UnregisteredSkill(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider()

	_, err := skills.Scan.Dispatch(context.Background(), provider, skills.ScanInput{
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
	if se.Phase != skills.PhaseDispatch {
		t.Fatalf("expected phase %q, got %q", skills.PhaseDispatch, se.Phase)
	}
}

func TestCollect_UnknownRef(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport())

	_, err := skills.Scan.Collect(context.Background(), provider, skills.JobRef{ID: "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown job ref")
	}

	var se *skills.SkillError
	if !errors.As(err, &se) {
		t.Fatalf("expected *skills.SkillError, got %T", err)
	}
	if se.Phase != skills.PhaseCollect {
		t.Fatalf("expected phase %q, got %q", skills.PhaseCollect, se.Phase)
	}
}

func TestCollect_CrossSkillRefRejected(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport()).
		WithTriageResult(skillstest.FixtureTriage())

	ctx := context.Background()
	scanRef, err := skills.Scan.Dispatch(ctx, provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Passing a Scan ref to Triage.Collect must fail.
	_, err = skills.Triage.Collect(ctx, provider, scanRef)
	if err == nil {
		t.Fatal("expected error when collecting with wrong skill's ref")
	}

	var se *skills.SkillError
	if !errors.As(err, &se) {
		t.Fatalf("expected *skills.SkillError, got %T", err)
	}
	if se.Phase != skills.PhaseCollect {
		t.Fatalf("expected phase %q, got %q", skills.PhaseCollect, se.Phase)
	}
}

func TestSyncAdapter_RoundTrip(t *testing.T) {
	asyncProvider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport())

	syncProvider := skills.SyncAdapter(asyncProvider)

	report, err := skills.Scan.Run(context.Background(), syncProvider, skills.ScanInput{
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

func TestSyncAdapter_ViaClient(t *testing.T) {
	asyncProvider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport()).
		WithTriageResult(skillstest.FixtureTriage())

	client := skills.NewClient(skills.SyncAdapter(asyncProvider))

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

func TestMultipleDispatches_IndependentRefs(t *testing.T) {
	provider := skillstest.NewStaticAsyncProvider().
		WithScanResult(skillstest.FixtureReport()).
		WithTriageResult(skillstest.FixtureTriage())

	ctx := context.Background()

	ref1, err := skills.Scan.Dispatch(ctx, provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		t.Fatal(err)
	}

	ref2, err := skills.Triage.Dispatch(ctx, provider, skills.TriageInput{
		Repo: "https://github.com/org/repo",
	})
	if err != nil {
		t.Fatal(err)
	}

	if ref1.ID == ref2.ID {
		t.Fatal("expected different job ref IDs for independent dispatches")
	}

	triage, err := skills.Triage.Collect(ctx, provider, ref2)
	if err != nil {
		t.Fatal(err)
	}
	if triage.TriageCompleted != "2026-01-01" {
		t.Fatalf("got triage completed %q, want %q", triage.TriageCompleted, "2026-01-01")
	}

	report, err := skills.Scan.Collect(ctx, provider, ref1)
	if err != nil {
		t.Fatal(err)
	}
	if report.Title != "Fixture Security Assessment" {
		t.Fatalf("got title %q, want %q", report.Title, "Fixture Security Assessment")
	}
}
