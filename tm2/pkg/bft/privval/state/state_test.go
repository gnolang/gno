package state

import (
	"testing"
)

func TestCommon_ValidateValidatorState(t *testing.T) {
	t.Parallel()

	// t.Run("valid validator state", func(t *testing.T) {
	// 	t.Parallel()
	//
	// 	fs := &FileState{}
	//
	// 	assert.NoError(t, fs.validate())
	// })
	//
	// t.Run("invalid step", func(t *testing.T) {
	// 	t.Parallel()
	//
	// 	fs := &FileState{}
	// 	fs.Step = StepPrecommit + 1 // invalid step
	//
	// 	assert.ErrorIs(t, fs.validate(), errInvalidSignStateStep)
	// })
	//
	// t.Run("invalid height", func(t *testing.T) {
	// 	t.Parallel()
	//
	// 	state := generateLastSignValidatorState()
	// 	state.Height = -1
	//
	// 	assert.ErrorIs(t, validateValidatorState(state), errInvalidSignStateHeight)
	// })
	//
	// t.Run("invalid round", func(t *testing.T) {
	// 	t.Parallel()
	//
	// 	state := generateLastSignValidatorState()
	// 	state.Round = -1
	//
	// 	assert.ErrorIs(t, validateValidatorState(state), errInvalidSignStateRound)
	// })
}
