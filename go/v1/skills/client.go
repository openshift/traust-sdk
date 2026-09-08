package skills

import (
	"context"

	"github.com/openshift/traust-sdk/go/v1/types"
)

// Client binds a Provider so callers don't pass it on every call. Use
// NewClient to create one, then call any skill method directly.
//
//	client := skills.NewClient(provider)
//	report, err := client.Scan(ctx, skills.ScanInput{Repo: url, Ref: "main"})
//	triage, err := client.Triage(ctx, skills.TriageInput{Repo: url})
type Client struct {
	p Provider
}

// NewClient creates a Client backed by the given Provider. All skill calls
// through this Client use the same Provider instance.
func NewClient(p Provider) *Client {
	return &Client{p: p}
}

// Scan runs the secure-code-audit skill. Returns a schema-validated Report.
func (c *Client) Scan(ctx context.Context, in ScanInput) (types.Report, error) {
	return Scan.Run(ctx, c.p, in)
}

// Triage runs the triage skill. Returns a schema-validated Triage.
func (c *Client) Triage(ctx context.Context, in TriageInput) (types.Triage, error) {
	return Triage.Run(ctx, c.p, in)
}

// VulnScan runs the vuln-scan skill. Returns a schema-validated VulnFindings.
func (c *Client) VulnScan(ctx context.Context, in VulnScanInput) (types.VulnFindings, error) {
	return VulnScan.Run(ctx, c.p, in)
}

// Validate runs the validate-findings skill. Returns a schema-validated Validation.
func (c *Client) Validate(ctx context.Context, in ValidateInput) (types.Validation, error) {
	return Validate.Run(ctx, c.p, in)
}

// Verify runs the verify-remediation skill. Returns a schema-validated Verification.
func (c *Client) Verify(ctx context.Context, in VerifyInput) (types.Verification, error) {
	return Verify.Run(ctx, c.p, in)
}

// Remediate runs the remediate-finding skill. Returns a schema-validated Remediation.
func (c *Client) Remediate(ctx context.Context, in RemediateInput) (types.Remediation, error) {
	return Remediate.Run(ctx, c.p, in)
}
