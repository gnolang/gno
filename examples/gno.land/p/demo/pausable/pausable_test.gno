package pausable

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
)

var (
	firstAddr  = std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	secondAddr = std.Address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa")
)

func TestNew(t *testing.T) {
	std.TestSetOrigCaller(firstAddr)

	result := New()

	if result.paused != false {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.paused)
	}

	if result.Owner() != firstAddr {
		t.Fatalf("Expected %s, got %s\n", firstAddr, result.Owner())
	}
}

func TestNewFromOwnable(t *testing.T) {
	std.TestSetOrigCaller(firstAddr)
	o := ownable.New()

	std.TestSetOrigCaller(secondAddr)
	result := NewFromOwnable(o)

	if result.Owner() != firstAddr {
		t.Fatalf("Expected %s, got %s\n", firstAddr, result.Owner())
	}
}

func TestSetUnpaused(t *testing.T) {
	std.TestSetOrigCaller(firstAddr)

	result := New()
	result.Unpause()

	if result.IsPaused() {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.IsPaused())
	}
}

func TestSetPaused(t *testing.T) {
	std.TestSetOrigCaller(firstAddr)

	result := New()
	result.Pause()

	if !result.IsPaused() {
		t.Fatalf("Expected result to be paused, got %t\n", result.IsPaused())
	}
}

func TestIsPaused(t *testing.T) {
	std.TestSetOrigCaller(firstAddr)

	result := New()

	if result.IsPaused() {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.IsPaused())
	}

	result.Pause()

	if !result.IsPaused() {
		t.Fatalf("Expected result to be paused, got %t\n", result.IsPaused())
	}
}
