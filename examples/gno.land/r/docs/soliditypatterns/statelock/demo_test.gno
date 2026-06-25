package statelock

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestLockUnlockHappyPath(cur realm, t *testing.T) {
	// Reset shared state before this test.
	state = StateActive

	uassert.Equal(t, string(state), string(StateActive), "initial state must be Active")

	Lock(cross(cur))
	uassert.Equal(t, string(state), string(StateLocked), "state must be Locked after Lock")

	Unlock(cross(cur))
	uassert.Equal(t, string(state), string(StateActive), "state must be Active after Unlock")

	// Restore state for subsequent tests.
	state = StateActive
}

func TestLockRejectsWrongState(cur realm, t *testing.T) {
	state = StateLocked

	uassert.AbortsWithMessage(t, cur, "invalid state: expected Active, got Locked", func() {
		Lock(cross(cur))
	}, "Lock must panic when state is not Active")

	// Restore state.
	state = StateActive
}

func TestUnlockRejectsWrongState(cur realm, t *testing.T) {
	state = StateActive

	uassert.AbortsWithMessage(t, cur, "invalid state: expected Locked, got Active", func() {
		Unlock(cross(cur))
	}, "Unlock must panic when state is not Locked")

	// state remains Active — no restore needed.
}
