// Package study owns source interpretation, evaluation, and semantic verification.
package study

import (
	"errors"

	"research-platform/seal"
)

type Evaluation struct {
	Outcome seal.TerminalOutcome
	Payload []byte // Study-owned JSON result, or failure record for TECHNICAL_FAILURE.
}

type Verification struct {
	Status string `json:"status"` // PASS or FAIL, including unfavorable complete results.
	Notes  string `json:"notes"`
}

// Evaluate reads local source bytes only after the platform has hashed the manifest inputs.
// Implement the frozen method here. Return INCOMPLETE for a valid partial result and
// TECHNICAL_FAILURE with a study-owned failure record for a terminal failure.
func Evaluate(in Inputs) (Evaluation, error) {
	return Evaluation{}, errors.New("implement study evaluation")
}

// Verify independently checks result semantics. Do not equate integrity with PASS.
func Verify(in Inputs, result []byte) (Verification, error) {
	return Verification{}, errors.New("implement study semantic verification")
}
