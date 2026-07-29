package commondao

import "testing"

// TestExecuteReentrancyLatch pins the latch mechanism directly: a nested
// enterExecute (which is what a re-entrant Execute reaches) must panic, and
// leaveExecute must clear the latch so the next execution can proceed. The
// end-to-end re-entry through a real proposal closure ships with the
// arbitrary-execution proposal kind (the closed catalog has no re-entering
// executor to drive it today).
func TestExecuteReentrancyLatch(t *testing.T) {
	leaveExecute() // ensure a clean starting state

	// First entry raises the latch.
	enterExecute()
	if !executing {
		t.Fatal("expected the latch to be raised after enterExecute")
	}

	// A nested entry (the re-entrant Execute) must panic.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected a nested enterExecute to panic")
			}
		}()
		enterExecute()
		t.Fatal("unreachable: the nested enterExecute must panic")
	}()

	// The rejected re-entry must not have lowered the latch.
	if !executing {
		t.Fatal("the latch must stay raised after a rejected re-entry")
	}

	// Leaving lowers it; the next execution may then proceed.
	leaveExecute()
	if executing {
		t.Fatal("expected the latch to be lowered after leaveExecute")
	}
	enterExecute()
	leaveExecute()
}
