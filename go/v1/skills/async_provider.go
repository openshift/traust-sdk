package skills

import "context"

// AsyncProvider is the async counterpart to [Provider]. Dispatch submits a
// skill and returns a [JobRef]; Collect blocks until the result is ready.
// The SDK validates inputs/outputs — the provider owns transport only.
type AsyncProvider interface {
	// Dispatch submits a skill and returns immediately with a handle.
	// Input []byte is pre-validated JSON.
	Dispatch(ctx context.Context, skill SkillMeta, input []byte) (JobRef, error)

	// Collect blocks until the result for ref is available or ctx cancels.
	// Returned []byte must be valid JSON conforming to the skill's output schema.
	Collect(ctx context.Context, ref JobRef) ([]byte, error)
}

// JobRef is an opaque handle correlating a Dispatch to its Collect result.
type JobRef struct {
	ID        string            // provider-assigned identifier, opaque to the SDK
	SkillName string            // skill that produced this ref; checked by Collect to prevent cross-skill misuse
	Metadata  map[string]string // provider-specific context carried between Dispatch and Collect
}

type syncAdapter struct {
	ap AsyncProvider
}

// SyncAdapter adapts an [AsyncProvider] into a synchronous [Provider].
// Each Execute call dispatches then immediately collects, blocking until done.
func SyncAdapter(ap AsyncProvider) Provider {
	return &syncAdapter{ap: ap}
}

func (a *syncAdapter) Execute(ctx context.Context, skill SkillMeta, input []byte) ([]byte, error) {
	ref, err := a.ap.Dispatch(ctx, skill, input)
	if err != nil {
		return nil, err
	}
	return a.ap.Collect(ctx, ref)
}
