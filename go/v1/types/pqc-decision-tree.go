// Code generated from traust-contracts v0.1.0. DO NOT EDIT.

package types

type PqcDecisionTree struct {
	FipsInteraction      PqcDecisionTreeFipsInteraction      `json:"fips_interaction"`
	Plan                 *string                             `json:"plan,omitempty"`
	PqcClassificationMap PqcDecisionTreePqcClassificationMap `json:"pqc_classification_map"`
	ProvenanceTree       PqcDecisionTreeProvenanceTree       `json:"provenance_tree"`
	ReadinessBuckets     PqcDecisionTreeReadinessBuckets     `json:"readiness_buckets"`
	RemediationEffort    PqcDecisionTreeRemediationEffort    `json:"remediation_effort"`
	Schema               *string                             `json:"schema,omitempty"`
	ServerSideCaveat     *PqcDecisionTreeServerSideCaveat    `json:"server_side_caveat,omitempty"`
	TlsControlCrosswalk  PqcDecisionTreeTlsControlCrosswalk  `json:"tls_control_crosswalk"`
	TreeVersion          string                              `json:"tree_version"`
}

type PqcDecisionTreeFipsInteraction struct {
	Comment *string                                   `json:"_comment,omitempty"`
	Rules   []PqcDecisionTreeFipsInteractionRulesItem `json:"rules"`
}

type PqcDecisionTreeFipsInteractionRulesItem struct {
	TreatAs *string `json:"treat_as,omitempty"`
	Verdict string  `json:"verdict"`
	When    string  `json:"when"`
}

type PqcDecisionTreePqcClassificationMap struct {
	Comment *string `json:"_comment,omitempty"`
}

type PqcDecisionTreeProvenanceTree struct {
	Comment *string                                  `json:"_comment,omitempty"`
	Rules   []PqcDecisionTreeProvenanceTreeRulesItem `json:"rules"`
}

type PqcDecisionTreeProvenanceTreeRulesItem struct {
	Provenance string `json:"provenance"`
	When       string `json:"when"`
}

type PqcDecisionTreeReadinessBuckets struct {
	Comment         *string                                    `json:"_comment,omitempty"`
	Buckets         interface{}                                `json:"buckets"`
	Rules           []PqcDecisionTreeReadinessBucketsRulesItem `json:"rules"`
	ScorePartialMin float64                                    `json:"score_partial_min"`
	ScoreReadyMin   float64                                    `json:"score_ready_min"`
}

type PqcDecisionTreeReadinessBucketsRulesItem struct {
	Bucket string `json:"bucket"`
	When   string `json:"when"`
}

type PqcDecisionTreeRemediationEffort struct {
	Comment *string                                     `json:"_comment,omitempty"`
	Classes interface{}                                 `json:"classes"`
	Rules   []PqcDecisionTreeRemediationEffortRulesItem `json:"rules"`
}

type PqcDecisionTreeRemediationEffortRulesItem struct {
	Effort string `json:"effort"`
	When   string `json:"when"`
}

type PqcDecisionTreeServerSideCaveat struct {
	Comment *string                                    `json:"_comment,omitempty"`
	Rules   []PqcDecisionTreeServerSideCaveatRulesItem `json:"rules"`
}

type PqcDecisionTreeServerSideCaveatRulesItem struct {
	Comment        *string `json:"_comment,omitempty"`
	Caveat         string  `json:"caveat"`
	ScoringCeiling *string `json:"scoring_ceiling,omitempty"`
	TreatAs        *string `json:"treat_as,omitempty"`
	When           string  `json:"when"`
}

type PqcDecisionTreeTlsControlCrosswalk struct {
	Comment          *string                `json:"_comment,omitempty"`
	AppControlled    map[string]interface{} `json:"app-controlled"`
	InfraControlled  map[string]interface{} `json:"infra-controlled"`
	VendorControlled map[string]interface{} `json:"vendor-controlled"`
}
