// Code generated from traust-contracts v1.2.0. DO NOT EDIT.

package types

type DocVariance struct {
	Metadata DocVarianceMetadata      `json:"metadata"`
	Records  []DocVarianceRecordsItem `json:"records"`
}

type DocVarianceMetadata struct {
	Created        string  `json:"created"`
	HarnessVersion string  `json:"harness_version"`
	Repository     string  `json:"repository"`
	Updated        *string `json:"updated,omitempty"`
}

type DocVarianceRecordsItem struct {
	Claim           string                                   `json:"claim"`
	CodeEvidence    []DocVarianceRecordsItemCodeEvidenceItem `json:"code_evidence"`
	Disposition     string                                   `json:"disposition"`
	DispositionNote *string                                  `json:"disposition_note,omitempty"`
	FindingRefs     []string                                 `json:"finding_refs,omitempty"`
	Id              string                                   `json:"id"`
	Source          DocVarianceRecordsItemSource             `json:"source"`
	ThreatRefs      []string                                 `json:"threat_refs,omitempty"`
	Variance        *string                                  `json:"variance,omitempty"`
	VerifiedAgainst *string                                  `json:"verified_against,omitempty"`
	VerifiedAt      string                                   `json:"verified_at"`
}

type DocVarianceRecordsItemCodeEvidenceItem struct {
	Lines *string `json:"lines,omitempty"`
	Note  *string `json:"note,omitempty"`
	Path  string  `json:"path"`
	Ref   *string `json:"ref,omitempty"`
	Repo  string  `json:"repo"`
}

type DocVarianceRecordsItemSource struct {
	Guide       *string `json:"guide,omitempty"`
	ProductSlug string  `json:"product_slug"`
	Quote       *string `json:"quote,omitempty"`
	Url         string  `json:"url"`
	Version     string  `json:"version"`
}
