// Code generated from traust-contracts v0.1.0. DO NOT EDIT.

package types

type ComplianceScope struct {
	Boundaries map[string]ComplianceScopeBoundariesEntry `json:"boundaries"`
	Updated    string                                    `json:"updated"`
	Version    int                                       `json:"version"`
}

type ComplianceScopeBoundariesEntry struct {
	DeclaredAt         string                                            `json:"declared_at"`
	DeclaredBy         string                                            `json:"declared_by"`
	DeploymentEvidence *ComplianceScopeBoundariesEntryDeploymentEvidence `json:"deployment_evidence,omitempty"`
	Exclude            []ComplianceScopeBoundariesEntryExcludeItem       `json:"exclude,omitempty"`
	Frameworks         []string                                          `json:"frameworks"`
	Include            []ComplianceScopeBoundariesEntryIncludeItem       `json:"include,omitempty"`
	Notes              *string                                           `json:"notes,omitempty"`
	Product            *string                                           `json:"product,omitempty"`
	ResolvesVia        string                                            `json:"resolves_via"`
}

type ComplianceScopeBoundariesEntryDeploymentEvidence struct {
	Inventory string `json:"inventory"`
	Service   string `json:"service"`
}

type ComplianceScopeBoundariesEntryExcludeItem struct {
	Reason string `json:"reason"`
	Repo   string `json:"repo"`
}

type ComplianceScopeBoundariesEntryIncludeItem struct {
	Reason string `json:"reason"`
	Repo   string `json:"repo"`
}
