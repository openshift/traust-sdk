package skills

import (
	"context"

	"github.com/openshift/traust-sdk/go/v1/types"
)

// AsyncClient binds an [AsyncProvider] so callers don't pass it on every call.
// Mirrors [Client] but exposes Dispatch*/Collect* pairs.
type AsyncClient struct {
	p AsyncProvider
}

// NewAsyncClient creates an AsyncClient backed by the given [AsyncProvider].
func NewAsyncClient(p AsyncProvider) *AsyncClient {
	return &AsyncClient{p: p}
}

// Provider returns the underlying AsyncProvider.
func (c *AsyncClient) Provider() AsyncProvider { return c.p }

// DispatchScan submits a scan job and returns a [JobRef] for later collection.
func (c *AsyncClient) DispatchScan(ctx context.Context, in ScanInput) (JobRef, error) {
	return Scan.Dispatch(ctx, c.p, in)
}

// CollectScan waits for a previously dispatched scan job to complete.
func (c *AsyncClient) CollectScan(ctx context.Context, ref JobRef) (types.Report, error) {
	return Scan.Collect(ctx, c.p, ref)
}

// DispatchTriage submits a triage job and returns a [JobRef] for later collection.
func (c *AsyncClient) DispatchTriage(ctx context.Context, in TriageInput) (JobRef, error) {
	return Triage.Dispatch(ctx, c.p, in)
}

// CollectTriage waits for a previously dispatched triage job to complete.
func (c *AsyncClient) CollectTriage(ctx context.Context, ref JobRef) (types.Triage, error) {
	return Triage.Collect(ctx, c.p, ref)
}

// DispatchVulnScan submits a vuln-scan job and returns a [JobRef] for later collection.
func (c *AsyncClient) DispatchVulnScan(ctx context.Context, in VulnScanInput) (JobRef, error) {
	return VulnScan.Dispatch(ctx, c.p, in)
}

// CollectVulnScan waits for a previously dispatched vuln-scan job to complete.
func (c *AsyncClient) CollectVulnScan(ctx context.Context, ref JobRef) (types.VulnFindings, error) {
	return VulnScan.Collect(ctx, c.p, ref)
}

// DispatchValidate submits a validate-findings job and returns a [JobRef] for later collection.
func (c *AsyncClient) DispatchValidate(ctx context.Context, in ValidateInput) (JobRef, error) {
	return Validate.Dispatch(ctx, c.p, in)
}

// CollectValidate waits for a previously dispatched validate-findings job to complete.
func (c *AsyncClient) CollectValidate(ctx context.Context, ref JobRef) (types.Validation, error) {
	return Validate.Collect(ctx, c.p, ref)
}

// DispatchVerify submits a verify-remediation job and returns a [JobRef] for later collection.
func (c *AsyncClient) DispatchVerify(ctx context.Context, in VerifyInput) (JobRef, error) {
	return Verify.Dispatch(ctx, c.p, in)
}

// CollectVerify waits for a previously dispatched verify-remediation job to complete.
func (c *AsyncClient) CollectVerify(ctx context.Context, ref JobRef) (types.Verification, error) {
	return Verify.Collect(ctx, c.p, ref)
}

// DispatchRemediate submits a remediate-finding job and returns a [JobRef] for later collection.
func (c *AsyncClient) DispatchRemediate(ctx context.Context, in RemediateInput) (JobRef, error) {
	return Remediate.Dispatch(ctx, c.p, in)
}

// CollectRemediate waits for a previously dispatched remediate-finding job to complete.
func (c *AsyncClient) CollectRemediate(ctx context.Context, ref JobRef) (types.Remediation, error) {
	return Remediate.Collect(ctx, c.p, ref)
}
