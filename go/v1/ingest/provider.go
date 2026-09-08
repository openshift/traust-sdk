package ingest

import "context"

// IngestKind identifies the ledger service submission kind.
type IngestKind string

const (
	KindTriage       IngestKind = "triage"
	KindValidation   IngestKind = "validation"
	KindVerification IngestKind = "verification"
	KindCountersign  IngestKind = "countersign"
	KindSeverity     IngestKind = "severity"
)

// JSON Schema names embedded in the contracts validate package.
const (
	SchemaTriage       = "triage"
	SchemaValidation   = "validation"
	SchemaVerification = "verification"
)

// IngestMeta carries routing metadata that the Provider needs to dispatch a
// human-lane submission.
type IngestMeta struct {
	Kind             IngestKind
	ContractsVersion string
}

// Provider submits validated payloads to the platform and returns raw response
// bytes. Implementations (HTTP client, mock, etc.) live outside this package.
type Provider interface {
	Ingest(ctx context.Context, meta IngestMeta, payload []byte) ([]byte, error)
	Submit(ctx context.Context, layerID string, payload []byte) ([]byte, error)
	Post(ctx context.Context, path string, payload []byte) ([]byte, error)
}

// EventLaneKindStrings returns ingest kind strings routed to POST /v1/ledger/events.
func EventLaneKindStrings() []string {
	return []string{string(KindCountersign), string(KindSeverity)}
}
