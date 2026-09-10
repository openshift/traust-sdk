// Code generated from traust-contracts v0.1.0. DO NOT EDIT.

package types

type RiskRatingMethodology struct {
	Bands              []string                                `json:"bands"`
	BucketThresholds   RiskRatingMethodologyBucketThresholds   `json:"bucket_thresholds"`
	Documentation      *string                                 `json:"documentation,omitempty"`
	Fallback           RiskRatingMethodologyFallback           `json:"fallback"`
	ImpactFactors      RiskRatingMethodologyImpactFactors      `json:"impact_factors"`
	LikelihoodFactors  RiskRatingMethodologyLikelihoodFactors  `json:"likelihood_factors"`
	Matrix             RiskRatingMethodologyMatrix             `json:"matrix"`
	Methodology        interface{}                             `json:"methodology"`
	MethodologyVersion string                                  `json:"methodology_version"`
	Schema             *string                                 `json:"schema,omitempty"`
	Source             string                                  `json:"source"`
	ThreatIntelFactor  *RiskRatingMethodologyThreatIntelFactor `json:"threat_intel_factor,omitempty"`
}

type CiaMap struct {
	H map[string]interface{} `json:"H"`
	L map[string]interface{} `json:"L"`
	N map[string]interface{} `json:"N"`
}

type MatrixRow struct {
	HIGH   string `json:"HIGH"`
	LOW    string `json:"LOW"`
	MEDIUM string `json:"MEDIUM"`
}

type RiskRatingMethodologyBucketThresholds struct {
	HighMin   float64 `json:"high_min"`
	MediumMin float64 `json:"medium_min"`
}

type RiskRatingMethodologyFallback struct {
	DefaultLikelihood float64                                     `json:"default_likelihood"`
	SeverityImpact    RiskRatingMethodologyFallbackSeverityImpact `json:"severity_impact"`
}

type RiskRatingMethodologyFallbackSeverityImpact struct {
	Critical      map[string]interface{} `json:"critical"`
	High          map[string]interface{} `json:"high"`
	Informational map[string]interface{} `json:"informational"`
	Low           map[string]interface{} `json:"low"`
	Medium        map[string]interface{} `json:"medium"`
}

type RiskRatingMethodologyImpactFactors struct {
	A CiaMap `json:"A"`
	C CiaMap `json:"C"`
	I CiaMap `json:"I"`
}

type RiskRatingMethodologyLikelihoodFactors struct {
	AC map[string]map[string]interface{} `json:"AC"`
	AV map[string]map[string]interface{} `json:"AV"`
	PR map[string]map[string]interface{} `json:"PR"`
	UI map[string]map[string]interface{} `json:"UI"`
}

type RiskRatingMethodologyMatrix struct {
	HIGH   MatrixRow `json:"HIGH"`
	LOW    MatrixRow `json:"LOW"`
	MEDIUM MatrixRow `json:"MEDIUM"`
}

type RiskRatingMethodologyThreatIntelFactor struct {
	Description *string                                               `json:"description,omitempty"`
	EpssBands   []RiskRatingMethodologyThreatIntelFactorEpssBandsItem `json:"epss_bands"`
	KevScore    float64                                               `json:"kev_score"`
}

type RiskRatingMethodologyThreatIntelFactorEpssBandsItem struct {
	MinEpss float64 `json:"min_epss"`
	Score   float64 `json:"score"`
}
