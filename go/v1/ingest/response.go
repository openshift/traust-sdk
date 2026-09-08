package ingest

// SubmitResponse is the ledger service response for report and event submissions.
type SubmitResponse struct {
	ID         string   `json:"id"`
	Status     string   `json:"status"`
	EventCount *int     `json:"event_count,omitempty"`
	MerkleRoot *string  `json:"merkle_root,omitempty"`
	EventIDs   []string `json:"event_ids,omitempty"`
	QueueAdded int      `json:"queue_added,omitempty"`
}

// ResolveResponse is the ledger service response for needs_review resolution.
type ResolveResponse struct {
	Resolved bool   `json:"resolved"`
	Key      string `json:"key"`
}

// FingerprintResponse is the ledger service response for fingerprint stamping.
type FingerprintResponse struct {
	Findings     []map[string]interface{} `json:"findings"`
	StampedCount int                      `json:"stamped_count"`
}

// SignResponse is the ledger service response for layer signing.
type SignResponse struct {
	Status  string `json:"status"`
	Method  string `json:"method"`
	LayerID string `json:"layer_id"`
}
