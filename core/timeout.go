package core

import (
	"context"
	"time"
)

// timeout is a holder for timeout duration information (constant)
type timeout struct {
	initial time.Duration // the initial timeout duration
	delta   time.Duration // the delta for future timeouts
}

// calculateTimeout calculates a new timeout duration using
// the formula:
//
// timeout(r) = initTimeout + r * timeoutDelta
func (t timeout) calculateTimeout(round uint64) time.Duration {
	return t.initial + time.Duration(round)*t.delta
}

// scheduleTimeout schedules a state timeout to be executed
func (t *Tendermint) scheduleTimeout(
	ctx context.Context,
	timeout time.Duration,
	callback func(),
) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		select {
		case <-ctx.Done():
		case <-time.After(timeout):
			callback()
		}
	}()
}

// onTimeoutPropose executes the <PREVOTE, nil> step
// as a result of the propose step timer going off
//
// 57: Function OnTimeoutPropose(height, round) :
// 58: 	if height = hp ∧ round = roundP ∧ stepP = propose then
// 59: 		broadcast <PREVOTE, hP, roundP, nil>
// 60: 		stepP ← prevote
func (t *Tendermint) onTimeoutPropose(round uint64) {
	var (
		// TODO Evaluate if the round information is even required.
		// We cancel the top-level timeout context upon every round change,
		// so this condition that the round != currentRound will always be false.
		// Essentially, I believe the only param we do need to check is
		// the current state in the SM, since this method can be executed async when
		// the SM is in a different state
		currentRound = t.state.LoadRound()
		currentStep  = t.state.step.Load()
	)

	// Make sure the timeout context is still valid
	if currentRound != round || currentStep != propose {
		// Timeout context no longer valid, ignore
		return
	}

	// Build and broadcast the prevote message, with an ID of NIL
	t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(nil))
}
