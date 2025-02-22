// examples/counter/counter_test.gno
package counter_test

import (
	"testing"
	"std"
	"gno.land/p/demo/counter"
)

func TestCounter(t *testing.T) {
	// Initialize state
	std.SetState(counter.State{Count: 0})

	// Test Increment
	counter.Increment()
	t.AssertEqual(counter.GetCount(), 1)

	// Test Decrement
	counter.Decrement()
	t.AssertEqual(counter.GetCount(), 0)

	// Test Reset
	counter.Increment()
	counter.Reset()
	t.AssertEqual(counter.GetCount(), 0)
}