package core

import (
	"context"
	"time"
)

// getDefaultTimeoutMap returns the default timeout map
// for the Tendermint consensus engine
func getDefaultTimeoutMap() map[step]Timeout {
	return map[step]Timeout{
		propose: {
			Initial: 10 * time.Second,       // 10s
			Delta:   500 * time.Millisecond, // 0.5
		},
		prevote: {
			Initial: 10 * time.Second,       // 10s
			Delta:   500 * time.Millisecond, // 0.5
		},
		precommit: {
			Initial: 10 * time.Second,       // 10s
			Delta:   500 * time.Millisecond, // 0.5
		},
	}
}

// Timeout is a holder for timeout duration information (constant)
type Timeout struct {
	Initial time.Duration // the initial timeout duration
	Delta   time.Duration // the delta for future timeouts
}

// CalculateTimeout calculates a new timeout duration using
// the formula:
//
// timeout(r) = initTimeout + r * timeoutDelta
func (t Timeout) CalculateTimeout(round uint64) time.Duration {
	return t.Initial + time.Duration(round)*t.Delta
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
		currentRound = t.state.getRound()
		currentStep  = t.state.step.get()
	)

	// Make sure the timeout context is still valid
	if currentRound != round || currentStep != propose {
		// Timeout context no longer valid, ignore
		return
	}

	// Build and broadcast the prevote message, with an ID of NIL
	t.broadcast.BroadcastPrevote(t.buildPrevoteMessage(nil))
}

// onTimeoutPrevote executes the <PRECOMMIT, nil> step
// as a result of the prevote step timer going off
//
// 61: Function OnTimeoutPrevote(height, round) :
// 62: 	if height = hp ∧ round = roundP ∧ stepP = prevote then
// 63: 		broadcast <PRECOMMIT, hP, roundP, nil>
// 64: 		stepP ← precommit
func (t *Tendermint) onTimeoutPrevote(round uint64) {
	var (
		// TODO Evaluate if the round information is even required.
		// We cancel the top-level timeout context upon every round change,
		// so this condition that the round != currentRound will always be false.
		// Essentially, I believe the only param we do need to check is
		// the current state in the SM, since this method can be executed async when
		// the SM is in a different state
		currentRound = t.state.getRound()
		currentStep  = t.state.step.get()
	)

	// Make sure the timeout context is still valid
	if currentRound != round || currentStep != prevote {
		// Timeout context no longer valid, ignore
		return
	}

	// Build and broadcast the prevote message, with an ID of NIL
	t.broadcast.BroadcastPrecommit(t.buildPrecommitMessage(nil))
}

// onTimeoutPrecommit executes the round expiration protocol
// as a result of the precommit step timer going off
//
// 65: Function OnTimeoutPrecommit(height, round) :
// 66: 	if height = hp ∧ round = roundP then
// 67: 		StartRound(roundP + 1)
func (t *Tendermint) onTimeoutPrecommit(round uint64, expiredCh chan<- struct{}) {
	// TODO Evaluate if the round information is even required.
	// We cancel the top-level timeout context upon every round change,
	// so this condition that the round != currentRound will always be false
	currentRound := t.state.getRound()

	// Make sure the timeout context is still valid
	if currentRound != round {
		// Timeout context no longer valid, ignore
		return
	}

	// Signal that the round expired (no consensus reached)
	select {
	case expiredCh <- struct{}{}:
	default:
	}
}
