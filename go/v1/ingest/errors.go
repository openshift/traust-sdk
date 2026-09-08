package ingest

import (
	"fmt"

	"github.com/openshift/traust-sdk/go/v1/ingest/internal/check"
)

// Phase identifies where in the submission pipeline an error occurred.
type Phase string

const (
	PhaseMarshal     Phase = "marshal"
	PhaseValidate    Phase = "validate"
	PhaseFingerprint Phase = "fingerprint"
	PhaseConvert     Phase = "convert"
	PhaseIngest      Phase = "ingest"
	PhaseDecode      Phase = "decode"
)

// IngestError wraps pipeline errors with the operation kind and phase for
// structured diagnostics.
type IngestError struct {
	Kind  IngestKind
	Phase Phase
	Err   error
}

func (e *IngestError) Error() string {
	return fmt.Sprintf("ingest[%s] %s: %v", e.Kind, e.Phase, e.Err)
}

func (e *IngestError) Unwrap() error { return e.Err }

func ingestErr(kind IngestKind, phase Phase, err error) *IngestError {
	return &IngestError{Kind: kind, Phase: phase, Err: err}
}

// FingerprintError is re-exported from the internal check package so consumers
// can use errors.As against fingerprint validation failures.
type FingerprintError = check.FingerprintError
