package skills

// ScanInput is the input for the secure-code-audit skill.
//
// Repo is the repository URL (e.g. "https://github.com/org/repo").
// Ref is the git ref to audit (branch, tag, or commit SHA).
type ScanInput struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
}

// TriageInput is the input for the triage skill.
//
// Repo is the repository URL. FindingsPath optionally points to a pre-existing
// findings file to triage. VotesPerFinding controls the n-of-m voting depth
// (default is skill-determined).
type TriageInput struct {
	Repo            string `json:"repo"`
	FindingsPath    string `json:"findings_path,omitempty"`
	VotesPerFinding int    `json:"votes_per_finding,omitempty"`
}

// VulnScanInput is the input for the vuln-scan skill.
//
// Repo and Ref identify the target. Diff enables incremental scanning against
// the baseline (only changes since the last audit). BaselinePath optionally
// overrides auto-discovery of the existing audit report. Focus restricts the
// scan to a single area.
type VulnScanInput struct {
	Repo         string `json:"repo"`
	Ref          string `json:"ref"`
	Diff         bool   `json:"diff,omitempty"`
	BaselinePath string `json:"baseline_path,omitempty"`
	Focus        string `json:"focus,omitempty"`
}

// ValidateInput is the input for the validate-findings skill.
//
// Repo is the repository URL. FindingsSource is the path or shorthand
// (product/repo, slug-findings, file path) pointing to the static findings to
// validate live. ScopeContext is an optional kubeconfig context name for scope
// binding. Namespace restricts validation to specific namespace(s). DryRun
// stops after attack planning without executing probes.
type ValidateInput struct {
	Repo           string `json:"repo"`
	FindingsSource string `json:"findings_source"`
	ScopeContext   string `json:"scope_context,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	DryRun         bool   `json:"dry_run,omitempty"`
}

// VerifyInput is the input for the verify-remediation skill.
//
// Repo is the repository URL. AuditReport is the path to the original
// *-security-audit.json that contains the findings to verify. PatchedRef is
// the branch, commit, or PR URL containing the fix. FixRepo optionally
// specifies a different repository where the fix landed (cross-repo fixes).
type VerifyInput struct {
	Repo        string `json:"repo"`
	AuditReport string `json:"audit_report"`
	PatchedRef  string `json:"patched_ref"`
	FixRepo     string `json:"fix_repo,omitempty"`
}

// RemediateInput is the input for the remediate-finding skill.
//
// Repo is the repository URL. RemID is the remediation manifest identifier
// (format: "<repo>.<finding_id>"). FindingID optionally narrows to a specific
// finding when the manifest row covers multiple.
type RemediateInput struct {
	Repo      string `json:"repo"`
	RemID     string `json:"rem_id"`
	FindingID string `json:"finding_id,omitempty"`
}
