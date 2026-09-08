// Package skills provides a typed SDK for invoking security harness skills
// through a pluggable execution backend.
//
// # Overview
//
// The primary entry point is a pre-wired [Skill] variable ([Scan], [Triage],
// [VulnScan], [Validate], [Verify], [Remediate]). Each Skill carries its typed
// input/output and handles JSON Schema validation internally — consumers never
// interact with schemas directly.
//
// # Quick start
//
// Implement the [Provider] interface (one method), then call any skill:
//
//	provider := myK8sProvider(...)  // implements skills.Provider
//	report, err := skills.Scan.Run(ctx, provider, skills.ScanInput{
//	    Repo: "https://github.com/org/repo",
//	    Ref:  "main",
//	})
//	for _, f := range report.Findings {
//	    fmt.Println(f.ID, f.Severity, f.Title)
//	}
//
// # Client convenience
//
// Bind a [Provider] once with [NewClient] to avoid passing it on every call:
//
//	client := skills.NewClient(provider)
//	report, _ := client.Scan(ctx, skills.ScanInput{...})
//	triage, _ := client.Triage(ctx, skills.TriageInput{...})
//
// # Custom skills
//
// Use [NewSkill] to define skills not in the built-in catalog, as long as their
// output conforms to an existing contracts schema:
//
//	var MyAudit = skills.NewSkill[MyInput, types.Report]("my-custom-audit", "", "report")
//	result, err := MyAudit.Run(ctx, provider, MyInput{...})
//
// # Testing
//
// The [skillstest] subpackage provides test doubles and fixtures:
//
//	p := skillstest.NewStaticProvider().WithScanResult(skillstest.FixtureReport())
//	report, _ := skills.Scan.Run(ctx, p, skills.ScanInput{...})
//
// # Error handling
//
// All errors from [Skill.Run] are [*SkillError] with a typed [Phase] indicating
// where the failure occurred (marshal, validate input, execute, validate output,
// decode output). Use [errors.As] to inspect:
//
//	var se *skills.SkillError
//	if errors.As(err, &se) {
//	    log.Printf("skill %s failed at %s: %v", se.Skill, se.Phase, se.Err)
//	}
package skills
