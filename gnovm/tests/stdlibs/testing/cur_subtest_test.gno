package testing_test

import "testing"

// A crossing sub-test entered after the test crossed into a helper must keep
// working. t.Run seeds the sub-test with the TOP-LEVEL test's cur on purpose
// (testing.gno), while the live crossing frame is the helper's minted cur, so
// the VM's stale-capture check has to exempt the testing stdlib's own
// dispatch. Without that exemption this test panics and aborts the run.

func curSubtestTarget(cur realm) {
	if !cur.IsCurrent() {
		panic("sub-test cur is not current")
	}
}

func curSubtestHelper(cur realm, t *testing.T) {
	t.Run("sub", func(cur realm, t *testing.T) {
		curSubtestTarget(cur)
	})
}

func TestCurSubtestAfterCross(cur realm, t *testing.T) {
	curSubtestHelper(cross(cur), t)
}
