// Code generated from traust-contracts v1.2.0. DO NOT EDIT.

package types

type ComplianceMapping struct {
	Checks   []ComplianceMappingChecksItem   `json:"checks"`
	Controls []ComplianceMappingControlsItem `json:"controls"`
	Note     *string                         `json:"note,omitempty"`
	Version  int                             `json:"version"`
}

type Assertion struct {
	Expected interface{} `json:"expected,omitempty"`
	Operator string      `json:"operator"`
	Path     string      `json:"path"`
}

type ComplianceMappingChecksItem struct {
	AppliesWhen   *Assertion `json:"applies_when,omitempty"`
	Assertion     Assertion  `json:"assertion"`
	Collector     string     `json:"collector"`
	Description   *string    `json:"description,omitempty"`
	Fixtures      *string    `json:"fixtures,omitempty"`
	Id            string     `json:"id"`
	ParameterRefs []string   `json:"parameter_refs,omitempty"`
	Sampling      *string    `json:"sampling,omitempty"`
}

type ComplianceMappingControlsItem struct {
	Applicability   *Assertion `json:"applicability,omitempty"`
	Checks          []string   `json:"checks,omitempty"`
	Classification  string     `json:"classification"`
	ControlId       string     `json:"control_id"`
	Crosswalk       []string   `json:"crosswalk,omitempty"`
	EvidenceRequest *string    `json:"evidence_request,omitempty"`
	Framework       string     `json:"framework"`
	Note            *string    `json:"note,omitempty"`
	Parameters      []string   `json:"parameters,omitempty"`
}
