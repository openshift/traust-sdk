// Code generated from traust-contracts v0.1.0. DO NOT EDIT.

package types

type SlaPolicy struct {
	ClockStart      *string                           `json:"clock_start,omitempty"`
	PolicyName      string                            `json:"policy_name"`
	Profiles        map[string]SlaPolicyProfilesEntry `json:"profiles"`
	SeverityMapping map[string]interface{}            `json:"severity_mapping"`
	Source          SlaPolicySource                   `json:"source"`
}

type SlaPolicyProfilesEntry struct {
	CvssFloorDays *SlaPolicyProfilesEntryCvssFloorDays       `json:"cvss_floor_days,omitempty"`
	Default       *bool                                      `json:"default,omitempty"`
	Description   *string                                    `json:"description,omitempty"`
	Slas          map[string]SlaPolicyProfilesEntrySlasEntry `json:"slas"`
}

type SlaPolicyProfilesEntryCvssFloorDays struct {
	ResolveDays int     `json:"resolve_days"`
	Threshold   float64 `json:"threshold"`
}

type SlaPolicyProfilesEntrySlasEntry struct {
	AcknowledgeDays interface{} `json:"acknowledge_days,omitempty"`
	ResolveDays     interface{} `json:"resolve_days"`
}

type SlaPolicySource struct {
	Name      string  `json:"name"`
	Note      *string `json:"note,omitempty"`
	Retrieved string  `json:"retrieved"`
	Url       *string `json:"url,omitempty"`
}
