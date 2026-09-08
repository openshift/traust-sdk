// Package skillstest provides test doubles and fixtures for the skills package.
//
// Use [StaticProvider] as a drop-in [skills.Provider] that returns pre-registered
// responses. Use [FixtureReport] and [FixtureTriage] for minimal schema-valid
// instances of common output types.
//
//	provider := skillstest.NewStaticProvider().
//	    WithScanResult(skillstest.FixtureReport())
//	report, err := skills.Scan.Run(ctx, provider, skills.ScanInput{...})
package skillstest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openshift/traust-sdk/go/v1/enums"
	"github.com/openshift/traust-sdk/go/v1/skills"
	"github.com/openshift/traust-sdk/go/v1/types"
)

// StaticProvider is a test double that returns pre-registered results keyed by
// skill name. Satisfies [skills.Provider].
type StaticProvider struct {
	results map[string][]byte
}

// NewStaticProvider creates a StaticProvider with no results registered.
func NewStaticProvider() *StaticProvider {
	return &StaticProvider{results: make(map[string][]byte)}
}

// WithResult registers a raw JSON response for a given skill name.
// Prefer the typed With*Result helpers; use this for custom skills.
func (p *StaticProvider) WithResult(skillName string, v any) *StaticProvider {
	p.results[skillName] = mustMarshal(v, skillName)
	return p
}

// WithScanResult registers a Report as the response for the Scan skill.
func (p *StaticProvider) WithScanResult(r types.Report) *StaticProvider {
	return p.WithResult(skills.Scan.Meta.Name, r)
}

// WithTriageResult registers a Triage as the response for the Triage skill.
func (p *StaticProvider) WithTriageResult(t types.Triage) *StaticProvider {
	return p.WithResult(skills.Triage.Meta.Name, t)
}

// WithVulnScanResult registers a VulnFindings as the response for the VulnScan skill.
func (p *StaticProvider) WithVulnScanResult(v types.VulnFindings) *StaticProvider {
	return p.WithResult(skills.VulnScan.Meta.Name, v)
}

// WithValidateResult registers a Validation as the response for the Validate skill.
func (p *StaticProvider) WithValidateResult(v types.Validation) *StaticProvider {
	return p.WithResult(skills.Validate.Meta.Name, v)
}

// WithVerifyResult registers a Verification as the response for the Verify skill.
func (p *StaticProvider) WithVerifyResult(v types.Verification) *StaticProvider {
	return p.WithResult(skills.Verify.Meta.Name, v)
}

// WithRemediateResult registers a Remediation as the response for the Remediate skill.
func (p *StaticProvider) WithRemediateResult(r types.Remediation) *StaticProvider {
	return p.WithResult(skills.Remediate.Meta.Name, r)
}

// Execute satisfies [skills.Provider]. Returns an error if no result is registered.
func (p *StaticProvider) Execute(_ context.Context, skill skills.SkillMeta, _ []byte) ([]byte, error) {
	raw, ok := p.results[skill.Name]
	if !ok {
		return nil, fmt.Errorf("skillstest: no result registered for skill %q", skill.Name)
	}
	return raw, nil
}

// StaticAsyncProvider is a test double for [skills.AsyncProvider].
// Not safe for concurrent use from multiple goroutines.
type StaticAsyncProvider struct {
	results map[string][]byte
	nextID  int
	refs    map[string]string // ref ID -> skill name
}

func NewStaticAsyncProvider() *StaticAsyncProvider {
	return &StaticAsyncProvider{
		results: make(map[string][]byte),
		refs:    make(map[string]string),
	}
}

func (p *StaticAsyncProvider) WithResult(skillName string, v any) *StaticAsyncProvider {
	p.results[skillName] = mustMarshal(v, skillName)
	return p
}

func (p *StaticAsyncProvider) WithScanResult(r types.Report) *StaticAsyncProvider {
	return p.WithResult(skills.Scan.Meta.Name, r)
}

func (p *StaticAsyncProvider) WithTriageResult(t types.Triage) *StaticAsyncProvider {
	return p.WithResult(skills.Triage.Meta.Name, t)
}

func (p *StaticAsyncProvider) WithVulnScanResult(v types.VulnFindings) *StaticAsyncProvider {
	return p.WithResult(skills.VulnScan.Meta.Name, v)
}

func (p *StaticAsyncProvider) WithValidateResult(v types.Validation) *StaticAsyncProvider {
	return p.WithResult(skills.Validate.Meta.Name, v)
}

func (p *StaticAsyncProvider) WithVerifyResult(v types.Verification) *StaticAsyncProvider {
	return p.WithResult(skills.Verify.Meta.Name, v)
}

func (p *StaticAsyncProvider) WithRemediateResult(r types.Remediation) *StaticAsyncProvider {
	return p.WithResult(skills.Remediate.Meta.Name, r)
}

func (p *StaticAsyncProvider) Dispatch(_ context.Context, skill skills.SkillMeta, _ []byte) (skills.JobRef, error) {
	if _, ok := p.results[skill.Name]; !ok {
		return skills.JobRef{}, fmt.Errorf("skillstest: no result registered for skill %q", skill.Name)
	}
	p.nextID++
	id := fmt.Sprintf("test-job-%d", p.nextID)
	p.refs[id] = skill.Name
	return skills.JobRef{ID: id, Metadata: map[string]string{"skill": skill.Name}}, nil
}

func (p *StaticAsyncProvider) Collect(_ context.Context, ref skills.JobRef) ([]byte, error) {
	skillName, ok := p.refs[ref.ID]
	if !ok {
		return nil, fmt.Errorf("skillstest: unknown job ref %q", ref.ID)
	}
	raw, ok := p.results[skillName]
	if !ok {
		return nil, fmt.Errorf("skillstest: no result registered for skill %q", skillName)
	}
	return raw, nil
}

func mustMarshal(v any, name string) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("skillstest: marshal fixture for %s: %v", name, err))
	}
	return raw
}

// --- Fixtures ---

// FixtureReport returns a minimal schema-valid Report for tests.
func FixtureReport() types.Report {
	return types.Report{
		Title: "Fixture Security Assessment",
		Metadata: types.ReportMetadata{
			Date:  "2026-01-01",
			Scope: "fixture scope",
		},
		ExecutiveSummary: types.ExecutiveSummary{
			Prose: "Fixture executive summary long enough to satisfy the schema's minimum length.",
		},
		SeverityCriteria: []types.SeverityCriterion{
			{Level: enums.SeverityCritical, Definition: "Fixture critical severity definition."},
			{Level: enums.SeverityHigh, Definition: "Fixture high severity definition."},
			{Level: enums.SeverityMedium, Definition: "Fixture medium severity definition."},
			{Level: enums.SeverityLow, Definition: "Fixture low severity definition."},
		},
		Findings: []types.ReportFinding{},
		FindingsSummary: []types.SeverityCountEntry{
			{Severity: enums.SeverityCritical, FindingIds: []string{}},
			{Severity: enums.SeverityHigh, FindingIds: []string{}},
			{Severity: enums.SeverityMedium, FindingIds: []string{}},
			{Severity: enums.SeverityLow, FindingIds: []string{}},
		},
		RemediationRoadmap: []types.RoadmapItem{
			{Priority: "P1", Action: "Fixture remediation action.", Addresses: []string{"none"}},
		},
	}
}

// FixtureTriage returns a minimal schema-valid Triage for tests.
func FixtureTriage() types.Triage {
	return types.Triage{
		TriageCompleted: "2026-01-01",
		TriageContext: types.TriageContext{
			Environment:     "fixture test environment",
			VotesPerFinding: 1,
			Repo:            "https://github.com/org/repo",
			HarnessVersion:  "0.24.0",
		},
		Summary:  types.TriageSummary{},
		Findings: []types.TriageFinding{},
	}
}
