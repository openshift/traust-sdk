package check

import (
	"testing"
)

func TestFingerprints_AllPresent(t *testing.T) {
	payload := []byte(`{"report":{"findings":[{"id":"F-001","fingerprint":"abc123"},{"id":"F-002","fingerprint":"def456"}]}}`)
	if err := Fingerprints(payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFingerprints_MissingFingerprint(t *testing.T) {
	payload := []byte(`{"report":{"findings":[{"id":"F-001","fingerprint":"abc123"},{"id":"F-002","fingerprint":""}]}}`)
	err := Fingerprints(payload)
	if err == nil {
		t.Fatal("expected error for missing fingerprint")
	}
	fpErr, ok := err.(*FingerprintError)
	if !ok {
		t.Fatalf("expected *FingerprintError, got %T", err)
	}
	if fpErr.FindingID != "F-002" {
		t.Fatalf("expected FindingID=F-002, got %q", fpErr.FindingID)
	}
}

func TestFingerprints_NoFingerprintField(t *testing.T) {
	payload := []byte(`{"report":{"findings":[{"id":"F-001"}]}}`)
	err := Fingerprints(payload)
	if err == nil {
		t.Fatal("expected error for missing fingerprint field")
	}
}

func TestFingerprints_TopLevelFindings(t *testing.T) {
	payload := []byte(`{"findings":[{"id":"F-001","fingerprint":"abc"}]}`)
	if err := Fingerprints(payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFingerprints_NonObjectPayload(t *testing.T) {
	payload := []byte(`"just a string"`)
	if err := Fingerprints(payload); err != nil {
		t.Fatalf("unexpected error for non-object: %v", err)
	}
}

func TestFingerprints_NoFindings(t *testing.T) {
	payload := []byte(`{"inventory_item_id":"item-1"}`)
	if err := Fingerprints(payload); err != nil {
		t.Fatalf("unexpected error for payload without findings: %v", err)
	}
}
