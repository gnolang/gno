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
	expiredCh chan<- struct{},
) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		select {
		case <-ctx.Done():
		case <-time.After(timeout):
			// Signal that the state expired
			select {
			case expiredCh <- struct{}{}:
			default:
			}
		}
	}()
}
