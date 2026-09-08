package skills

import "fmt"

// Phase identifies where in the Skill.Run pipeline a failure occurred.
type Phase string

const (
	PhaseMarshalInput   Phase = "marshal_input"
	PhaseValidateInput  Phase = "validate_input"
	PhaseExecute        Phase = "execute"
	PhaseDispatch       Phase = "dispatch"
	PhaseCollect        Phase = "collect"
	PhaseValidateOutput Phase = "validate_output"
	PhaseDecodeOutput   Phase = "decode_output"
)

// SkillError is the structured error returned from Skill.Run. Consumers can
// type-assert or use errors.As to inspect which phase failed and which skill
// was being invoked.
//
//	var se *skills.SkillError
//	if errors.As(err, &se) {
//	    switch se.Phase {
//	    case skills.PhaseExecute:
//	        // provider-level failure — retry or escalate
//	    case skills.PhaseValidateOutput:
//	        // skill returned invalid data — likely a contracts version mismatch
//	    }
//	}
type SkillError struct {
	// Skill is the name of the skill that failed (e.g. "secure-code-audit").
	Skill string

	// Phase identifies where in the pipeline the failure occurred.
	Phase Phase

	// Err is the underlying cause.
	Err error
}

func (e *SkillError) Error() string {
	return fmt.Sprintf("skills[%s] %s: %v", e.Skill, e.Phase, e.Err)
}

func (e *SkillError) Unwrap() error {
	return e.Err
}

func skillErr(name string, phase Phase, err error) *SkillError {
	return &SkillError{Skill: name, Phase: phase, Err: err}
}
