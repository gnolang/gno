package main

import "testing"

func TestCounter_Increment(t *testing.T) {
	counter := Counter{value: 0}

	// Capture the initial value
	initialValue := counter.value

	// Call the Increment method
	counter.Increment()

	// Check if the value has changed (this should fail if using value receiver)
	if counter.value == initialValue {
		t.Fatalf("Expected value to be incremented, but it remained %d", counter.value)
	}
}
