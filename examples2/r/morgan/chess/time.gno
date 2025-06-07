package chess

import (
	"time"
)

// TimeControl keeps track of time control information for the game.
type TimeControl struct {
	Seconds   int
	Increment int

	StartedAt      time.Time
	MoveTimestamps []time.Time
	WhiteTime      time.Duration
	BlackTime      time.Duration
}

func NewTimeControl(seconds, incr int) *TimeControl {
	if seconds <= 0 {
		return nil
	}
	return &TimeControl{
		Seconds:   seconds,
		Increment: incr,

		StartedAt: time.Now(),
		WhiteTime: time.Duration(seconds) * time.Second,
		BlackTime: time.Duration(seconds) * time.Second,
	}
}

// AddMove records that at the current time, a new move was added.
func (tc *TimeControl) AddMove() (valid bool) {
	nd, v := tc.timedOut()
	if v {
		return false
	}
	if len(tc.MoveTimestamps)&1 == 0 {
		tc.WhiteTime = nd
	} else {
		tc.BlackTime = nd
	}
	tc.MoveTimestamps = append(tc.MoveTimestamps, time.Now())
	return true
}

func (tc *TimeControl) TimedOut() bool {
	_, v := tc.timedOut()
	return v
}

func (tc *TimeControl) timedOut() (time.Duration, bool) {
	mts := tc.MoveTimestamps

	// First move for each player: they both have up to 30 seconds.
	switch len(mts) {
	case 0:
		delta := time.Since(tc.StartedAt)
		return tc.WhiteTime, delta > time.Second*30
	case 1:
		delta := time.Since(mts[0])
		return tc.BlackTime, delta > time.Second*30
	}

	// Determine color. Determine time since last move. Try subtracting from
	// color's time. If >= 0, good. If < 0, timeout.
	delta := time.Since(mts[len(mts)-1])

	if len(mts)&1 == 0 { // white
		nt := tc.WhiteTime - delta
		return nt + tc.incr(), nt < 0
	}

	nt := tc.BlackTime - delta
	return nt + tc.incr(), nt < 0
}

func (tc *TimeControl) incr() time.Duration {
	// there is always at least a one second increment, to account for
	// block time and the delay between user making a move and tx happening
	return time.Second + time.Duration(tc.Increment)*time.Second
}
