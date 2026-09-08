package check

import (
	"encoding/json"
	"fmt"
)

// FingerprintError indicates a finding is missing its fingerprint.
type FingerprintError struct {
	FindingID string
}

func (e *FingerprintError) Error() string {
	return fmt.Sprintf("finding %q missing fingerprint", e.FindingID)
}

// Fingerprints checks that every finding in the payload has a non-empty
// fingerprint field. It never recomputes fingerprints — the harness skill is
// the sole producer.
//
// The check inspects known finding containers: top-level "findings" arrays and
// the "report.findings" path used by ledger service report payloads.
func Fingerprints(payload []byte) error {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(payload, &doc); err != nil {
		return nil // non-object payloads don't have findings
	}

	if raw, ok := doc["report"]; ok {
		if err := checkReportFindings(raw); err != nil {
			return err
		}
	}

	if raw, ok := doc["findings"]; ok {
		return checkFindingsArray(raw)
	}

	return nil
}

func checkReportFindings(raw json.RawMessage) error {
	var report struct {
		Findings []findingStub `json:"findings"`
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil
	}
	for _, f := range report.Findings {
		if f.Fingerprint == "" {
			return &FingerprintError{FindingID: f.ID}
		}
	}
	return nil
}

func checkFindingsArray(raw json.RawMessage) error {
	var findings []findingStub
	if err := json.Unmarshal(raw, &findings); err != nil {
		return nil
	}
	for _, f := range findings {
		if f.Fingerprint == "" {
			return &FingerprintError{FindingID: f.ID}
		}
	}
	return nil
}

type findingStub struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
}

// EventFingerprints checks that every derived event carries a non-empty
// fingerprint. This is the correct validation point for report lanes (triage,
// validation, verification): those reports do not carry fingerprints on their
// findings — the fingerprint is a scan-lane identity supplied by the caller and
// stamped onto events during conversion. Checking the raw report would always
// fail; checking the converted events verifies the caller supplied a
// fingerprint for every finding that produced an event.
//
// The finding_ref (not the fingerprint) is reported when one is missing so the
// error names the offending finding.
func EventFingerprints(events []map[string]interface{}) error {
	for _, ev := range events {
		fp, _ := ev["fingerprint"].(string)
		if fp == "" {
			ref, _ := ev["finding_ref"].(string)
			return &FingerprintError{FindingID: ref}
		}
	}
	return nil
}
