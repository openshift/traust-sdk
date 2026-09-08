# traust-sdk (Go)

```bash
go get github.com/openshift/traust-sdk/go@v0.9.0
```

## Skills SDK

Import a skill, plug in your provider, call it — typed input in, typed result out:

```go
import "github.com/openshift/traust-sdk/go/v1/skills"

provider := myK8sProvider(...)  // you implement skills.Provider

report, err := skills.Scan.Run(ctx, provider, skills.ScanInput{
    Repo: "https://github.com/org/repo",
    Ref:  "main",
})
for _, f := range report.Findings {
    fmt.Println(f.ID, f.Severity, f.Title)
}
```

`skills.Provider` is the only interface you implement — it runs a skill somewhere
(K8s Job, Docker, local subprocess, ...) and returns raw bytes:

```go
type Provider interface {
    Execute(ctx context.Context, skill SkillMeta, input []byte) ([]byte, error)
}
```

The SDK handles input/output schema validation, marshaling, and type decoding.

### Calling multiple skills

Bind the provider once with a Client:

```go
client := skills.NewClient(provider)
report, _ := client.Scan(ctx, skills.ScanInput{...})
triage, _ := client.Triage(ctx, skills.TriageInput{...})
```

### Testing

```go
import "github.com/openshift/traust-sdk/go/v1/skills/skillstest"

provider := skillstest.NewStaticProvider().
    WithScanResult(skillstest.FixtureReport())

report, _ := skills.Scan.Run(ctx, provider, skills.ScanInput{
    Repo: "https://github.com/org/repo",
    Ref:  "main",
})
```

## Ingest SDK

Submit reports and human events to the ledger service:

```go
import "github.com/openshift/traust-sdk/go/v1/ingest"

client := ingest.NewHTTPClient("https://ledger.example.com",
    ingest.WithBearerToken(os.Getenv("LEDGER_TOKEN")),
)

// Machine lane — triage report (derives disposition events)
resp, err := client.SubmitTriageReport(ctx, ingest.TriageReportInput{
    ReportMeta: ingest.ReportMeta{
        LayerID:    "repo-a",
        SourceRef:  "findings/repo-a/triage.json",
        RecordedAt: time.Now().UTC().Format(time.RFC3339),
    },
    Report: triageReport,
})

// Human lane — countersign
resp, err = client.SubmitCountersign(ctx, ingest.CountersignInput{
    EventMeta: ingest.EventMeta{
        LayerID:    "repo-a",
        RecordedAt: time.Now().UTC().Format(time.RFC3339),
    },
    FindingRef:    "FIND-001",
    Verdict:       "true_positive",
    Justification: "Reviewed source and confirmed exploit path.",
})

// Sign a layer's merkle tree
signResp, err := client.SignLayer(ctx, "repo-a", ingest.SignOpts{})
```

Machine-lane report kinds: `triage`, `validation`. Human-lane event kinds:
`countersign`, `severity`. Additional operations: `BatchSubmit`,
`ResolveReviewItem`, `ComputeFingerprints`, `SignLayer`.

The built-in HTTP transport handles envelope formatting, lane routing, and auth.
For custom transports, implement `ingest.Provider` and use `ingest.NewClient(provider)`.

### Testing

```go
import "github.com/openshift/traust-sdk/go/v1/ingest/ingesttest"

provider := ingesttest.NewStaticProvider().
    WithTriageResponse(ingesttest.FixtureTriageResponse())
client := ingest.NewClient(provider)
```

## Query SDK

Read layers, findings, and verification results from the ledger service:

```go
import "github.com/openshift/traust-sdk/go/v1/query"

client := query.NewHTTPClient("https://ledger.example.com",
    query.WithBearerToken(os.Getenv("LEDGER_TOKEN")),
)

// Read resolved findings for a layer
findings, err := client.GetFindings(ctx, "repo-a")
// Paginated bulk findings
page, err := client.ListFindings(ctx, query.ListFindingsOpts{Limit: 50})
// Verify layer integrity
result, err := client.VerifyLayer(ctx, "repo-a", query.VerifyOpts{CheckSignatures: true})
```

The built-in HTTP transport handles auth and path construction.
For custom transports, implement `query.Provider` and use `query.NewClient(provider)`.

### Testing

```go
import "github.com/openshift/traust-sdk/go/v1/query/querytest"

provider := querytest.NewStaticProvider().
    WithFindingsResponse("repo-a", querytest.FixtureFindingsResponse())
client := query.NewClient(provider)
```

## Data-only usage

For consumers who just need types or validation without invoking skills:

| Package | Contents |
|---|---|
| [`v1/types`](v1/types) | Generated Go structs for every schema in `schemas/v1` |
| [`v1/enums`](v1/enums) | Generated shared vocabularies (severity, validity, verdicts, ...) |
| [`v1/validate`](v1/validate) | Validate raw JSON bytes against an embedded v1 schema |

## Regenerating types/enums/validate

```bash
make generate     # reads schemas/v1 + enums/v1, writes v1/{types,enums,validate}
make check-drift  # CI: fails if generated output is stale
```
