package ingest

import "context"

// Client binds a Provider for typed submission methods.
type Client struct {
	p Provider
}

// NewClient creates a Client backed by the given Provider.
func NewClient(p Provider) *Client { return &Client{p: p} }

// SubmitTriageReport converts a triage report to events and batch-submits them.
func (c *Client) SubmitTriageReport(ctx context.Context, in TriageReportInput) (SubmitResponse, error) {
	return triageOp.Submit(ctx, c.p, in)
}

// SubmitValidationReport converts a validation report to events and batch-submits them.
func (c *Client) SubmitValidationReport(ctx context.Context, in ValidationReportInput) (SubmitResponse, error) {
	return validationOp.Submit(ctx, c.p, in)
}

// SubmitVerificationReport converts a verification report to events and batch-submits them.
func (c *Client) SubmitVerificationReport(ctx context.Context, in VerificationReportInput) (SubmitResponse, error) {
	return verificationOp.Submit(ctx, c.p, in)
}

// BatchSubmit sends pre-formed events via POST /v1/ledger/layers/{layer_id}/submit.
func (c *Client) BatchSubmit(ctx context.Context, layerID string, input BatchSubmitInput) (SubmitResponse, error) {
	return BatchSubmit(ctx, c.p, layerID, input)
}

// ResolveReviewItem resolves a needs_review queue item.
func (c *Client) ResolveReviewItem(ctx context.Context, layerID string, input ResolveInput) (ResolveResponse, error) {
	return ResolveReviewItem(ctx, c.p, layerID, input)
}

// ComputeFingerprints stamps fingerprints on findings.
func (c *Client) ComputeFingerprints(ctx context.Context, input FingerprintInput) (FingerprintResponse, error) {
	return ComputeFingerprints(ctx, c.p, input)
}

// SignLayer requests the ledger service to sign a layer's merkle tree.
func (c *Client) SignLayer(ctx context.Context, layerID string, opts SignOpts) (SignResponse, error) {
	return SignLayer(ctx, c.p, layerID, opts)
}

// SubmitCountersign submits a human countersign decision for a finding.
func (c *Client) SubmitCountersign(ctx context.Context, in CountersignInput) (SubmitResponse, error) {
	return countersignOp.Submit(ctx, c.p, in)
}

// SubmitSeverity submits a human severity override for a finding.
func (c *Client) SubmitSeverity(ctx context.Context, in SeverityInput) (SubmitResponse, error) {
	return severityOp.Submit(ctx, c.p, in)
}
