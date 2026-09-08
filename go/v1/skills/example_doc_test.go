package skills_test

import (
	"context"
	"fmt"

	"github.com/openshift/traust-sdk/go/v1/skills"
	"github.com/openshift/traust-sdk/go/v1/skills/skillstest"
)

// This example shows the primary usage pattern: call a skill directly with a
// Provider.
func Example_scanDirect() {
	provider := skillstest.NewStaticProvider().
		WithScanResult(skillstest.FixtureReport())

	report, err := skills.Scan.Run(context.Background(), provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(report.Title)
	// Output: Fixture Security Assessment
}

// This example shows binding a Provider once with a Client for multiple skill
// calls.
func ExampleClient_Scan() {
	provider := skillstest.NewStaticProvider().
		WithScanResult(skillstest.FixtureReport())
	client := skills.NewClient(provider)

	report, err := client.Scan(context.Background(), skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(report.Title)
	// Output: Fixture Security Assessment
}

// This example shows the test double pattern with StaticProvider.
func ExampleNewStaticProvider() {
	provider := skillstest.NewStaticProvider().
		WithScanResult(skillstest.FixtureReport()).
		WithTriageResult(skillstest.FixtureTriage())

	report, _ := skills.Scan.Run(context.Background(), provider, skills.ScanInput{
		Repo: "https://github.com/org/repo",
		Ref:  "main",
	})
	fmt.Println(report.Title)

	triage, _ := skills.Triage.Run(context.Background(), provider, skills.TriageInput{
		Repo: "https://github.com/org/repo",
	})
	fmt.Println(triage.TriageContext.HarnessVersion)
	// Output:
	// Fixture Security Assessment
	// 0.24.0
}

// This example shows defining a custom skill with NewSkill.
func ExampleNewSkill() {
	type CustomInput struct {
		Target string `json:"target"`
	}

	customScan := skills.NewSkill[CustomInput, any]("my-custom-scan", "", "report")
	fmt.Println(customScan.Meta.Name)
	// Output: my-custom-scan
}
