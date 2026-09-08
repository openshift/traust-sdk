package skills

import "context"

// SkillMeta identifies a skill invocation to the Provider. It carries the skill
// name and the contracts version so the Provider can route or version-stamp as
// needed without importing schema packages.
type SkillMeta struct {
	// Name is the harness skill identifier (e.g. NameScan, NameTriage).
	Name string

	// ContractsVersion is the traust-contracts semver these types were
	// generated from (e.g. "0.3.0").
	ContractsVersion string
}

// Provider runs a skill somewhere and returns the raw result bytes.
//
// Implementations (K8s Job, Docker container, local subprocess, etc.) live
// outside this package. The SDK owns input/output validation; Provider owns
// only execution transport.
//
// The input []byte is the JSON-marshaled skill input. The returned []byte must
// be valid JSON conforming to the skill's output schema.
type Provider interface {
	Execute(ctx context.Context, skill SkillMeta, input []byte) ([]byte, error)
}
