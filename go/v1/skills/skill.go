package skills

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openshift/traust-sdk/go/v1/types"
	"github.com/openshift/traust-sdk/go/v1/validate"
)

// Skill is a typed, schema-validated invocation unit. Each Skill knows its
// input and output Go types plus the JSON Schema names used for validation.
//
// Consumers do not construct Skills — use the pre-wired package-level variables
// (Scan, Triage, VulnScan, Validate, Verify, Remediate).
type Skill[In, Out any] struct {
	// Meta carries the skill name and contracts version for the Provider.
	Meta         SkillMeta
	inputSchema  string
	outputSchema string
}

// NewSkill creates a custom Skill with the given name and output schema. The
// output schema name must match a schema embedded in the contracts validate
// package (e.g. "report", "triage", "validation"). Input schema is optional —
// pass "" to skip input validation.
//
// Use this when you have a custom harness skill whose output conforms to an
// existing contracts schema but isn't in the SDK's built-in catalog.
//
//	var MySkill = skills.NewSkill[MyInput, types.Report]("my-custom-audit", "", "report")
func NewSkill[In, Out any](name, inputSchema, outputSchema string) Skill[In, Out] {
	return Skill[In, Out]{
		Meta:         SkillMeta{Name: name, ContractsVersion: types.ContractsVersion},
		inputSchema:  inputSchema,
		outputSchema: outputSchema,
	}
}

// Run executes the skill against a Provider:
//
//  1. Marshal the typed input to JSON.
//  2. Validate input against the input schema (if one is defined for this skill).
//  3. Call p.Execute to run the skill on whatever backend the Provider wraps.
//  4. Validate the raw output against the output schema.
//  5. Unmarshal the output into the typed result.
//
// Returns the fully typed, schema-validated result or an error at any stage.
func (s Skill[In, Out]) Run(ctx context.Context, p Provider, in In) (Out, error) {
	var zero Out

	input, err := json.Marshal(in)
	if err != nil {
		return zero, skillErr(s.Meta.Name, PhaseMarshalInput, err)
	}

	if s.inputSchema != "" {
		if err := validate.ValidateBytes(s.inputSchema, input); err != nil {
			return zero, skillErr(s.Meta.Name, PhaseValidateInput, err)
		}
	}

	raw, err := p.Execute(ctx, s.Meta, input)
	if err != nil {
		return zero, skillErr(s.Meta.Name, PhaseExecute, err)
	}

	if err := validate.ValidateBytes(s.outputSchema, raw); err != nil {
		return zero, skillErr(s.Meta.Name, PhaseValidateOutput, err)
	}

	var out Out
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, skillErr(s.Meta.Name, PhaseDecodeOutput, err)
	}
	return out, nil
}

// Dispatch validates and submits the skill input via an [AsyncProvider],
// returning a [JobRef] handle. Call [Skill.Collect] to retrieve the result.
func (s Skill[In, Out]) Dispatch(ctx context.Context, p AsyncProvider, in In) (JobRef, error) {
	input, err := json.Marshal(in)
	if err != nil {
		return JobRef{}, skillErr(s.Meta.Name, PhaseMarshalInput, err)
	}

	if s.inputSchema != "" {
		if err := validate.ValidateBytes(s.inputSchema, input); err != nil {
			return JobRef{}, skillErr(s.Meta.Name, PhaseValidateInput, err)
		}
	}

	ref, err := p.Dispatch(ctx, s.Meta, input)
	if err != nil {
		return JobRef{}, skillErr(s.Meta.Name, PhaseDispatch, err)
	}
	ref.SkillName = s.Meta.Name
	return ref, nil
}

// Collect retrieves and validates the result for a previously dispatched ref.
func (s Skill[In, Out]) Collect(ctx context.Context, p AsyncProvider, ref JobRef) (Out, error) {
	var zero Out

	if ref.SkillName != "" && ref.SkillName != s.Meta.Name {
		return zero, skillErr(s.Meta.Name, PhaseCollect,
			fmt.Errorf("ref was dispatched for skill %q, not %q", ref.SkillName, s.Meta.Name))
	}

	raw, err := p.Collect(ctx, ref)
	if err != nil {
		return zero, skillErr(s.Meta.Name, PhaseCollect, err)
	}

	if err := validate.ValidateBytes(s.outputSchema, raw); err != nil {
		return zero, skillErr(s.Meta.Name, PhaseValidateOutput, err)
	}

	var out Out
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, skillErr(s.Meta.Name, PhaseDecodeOutput, err)
	}
	return out, nil
}

// Scan runs the secure-code-audit skill.
// Input: ScanInput (repo URL + git ref).
// Output: types.Report — a full security audit report with findings, severity
// criteria, remediation roadmap, and executive summary.
var Scan = NewSkill[ScanInput, types.Report](SkillSecureCodeAudit, "", "report")

// Triage runs the triage skill.
// Input: TriageInput (repo URL + optional findings path and vote count).
// Output: types.Triage — deduplicated, verdict-scored findings with confidence
// and triage context.
var Triage = NewSkill[TriageInput, types.Triage](SkillTriage, "", "triage")

// VulnScan runs the vuln-scan skill.
// Input: VulnScanInput (repo + ref, optional diff mode and baseline).
// Output: types.VulnFindings — fast vulnerability candidates for triage,
// lighter than a full secure-code-audit.
var VulnScan = NewSkill[VulnScanInput, types.VulnFindings](SkillVulnScan, "", "vuln-findings")

// Validate runs the validate-findings skill.
// Input: ValidateInput (repo + findings source + scope binding).
// Output: types.Validation — live-tested attack chains proving or refuting
// static findings against a real environment.
var Validate = NewSkill[ValidateInput, types.Validation](SkillValidateFindings, "", "validation")

// Verify runs the verify-remediation skill.
// Input: VerifyInput (repo + original audit report + patched ref).
// Output: types.Verification — per-finding verdicts on whether fixes resolve
// the original vulnerabilities, with regression detection.
var Verify = NewSkill[VerifyInput, types.Verification](SkillVerifyRemedation, "", "verification")

// Remediate runs the remediate-finding skill.
// Input: RemediateInput (repo + remediation manifest ID).
// Output: types.Remediation — patch metadata, check results, PR info, and
// revalidation status for an automated security fix.
var Remediate = NewSkill[RemediateInput, types.Remediation](SkillRemediateFinding, "", "remediation")
