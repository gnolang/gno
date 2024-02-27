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

// scheduleTimeoutPropose schedules a future timeout propose trigger
func (t *Tendermint) scheduleTimeoutPropose(ctx context.Context) {
	// TODO Make thread safe
	// Fetch the current view, before the trigger is set
	var (
		round = t.state.view.Round

		timeoutPropose = t.timeouts[t.state.step].calculateTimeout(round)
	)

	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		select {
		case <-ctx.Done():
		case <-time.After(timeoutPropose):
			t.onTimeoutPropose(round)
		}
	}()
}

// onTimeoutPropose executes the <PREVOTE, nil> step
// as a result of the propose step timer going off
func (t *Tendermint) onTimeoutPropose(round uint64) {
	// TODO make thread safe
	var (
		// TODO Evaluate if the round information is even required.
		// We cancel the top-level timeout context upon every round change,
		// so this condition that the round != currentRound will always be false.
		// Essentially, I believe the only param we do need to check is
		// the current state in the SM, since this method can be executed async when
		// the SM is in a different state
		currentRound = t.state.view.Round
		currentStep  = t.state.step
	)

	// Make sure the timeout context is still valid
	if currentRound != round || currentStep != propose {
		// Timeout context no longer valid, ignore
		return
	}

	// Build and broadcast the prevote message, with an ID of NIL
	t.broadcastPrevote(t.buildPrevoteMessage(nil))
}
