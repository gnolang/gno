// examples/counter/counter.gno
package counter

import (
	"std"
)

// State holds the counter value.
type State struct {
	Count int
}

// Increment increases the counter by 1.
func Increment() {
	state := std.GetState().(State) // Load state
	state.Count++
	std.SetState(state)            // Save state
}

// Decrement decreases the counter by 1.
func Decrement() {
	state := std.GetState().(State)
	state.Count--
	std.SetState(state)
}

// GetCount returns the current counter value.
func GetCount() int {
	state := std.GetState().(State)
	return state.Count
}

// Reset sets the counter to 0.
func Reset() {
	std.SetState(State{Count: 0})
}