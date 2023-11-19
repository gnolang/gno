// counter-app/r/counter/counter_test.gno

package counter

import "testing"

func TestCounter_Increment(t *testing.T) {
	// Reset the value
	count = 0

	// Verify the initial value is 0
	if count != 0 {
		t.Fatalf("initial value != 0")
	}

	// Increment the value
	Increment()

	// Verify the initial value is 1
	if count != 1 {
		t.Fatalf("initial value != 1")
	}
}

func TestCounter_Decrement(t *testing.T) {
	// Reset the value
	count = 0

	// Verify the initial value is 0
	if count != 0 {
		t.Fatalf("initial value != 0")
	}

	// Decrement the value
	Decrement()

	// Verify the initial value is 1
	if count != -1 {
		t.Fatalf("initial value != -1")
	}
}

func TestCounter_Render(t *testing.T) {
	// Reset the value
	count = 0

	// Verify the Render output
	if Render("") != "Count: 0" {
		t.Fatalf("invalid Render value")
	}
}
