package pausable

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
)

var (
	firstCaller  = std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")
	secondCaller = std.Address("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	result := New()

	if result.paused != false {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.paused)
	}

	if result.Owner() != firstCaller {
		t.Fatalf("Expected %s, got %s\n", firstCaller, result.Owner())
	}
}

func TestNewFromOwnable(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	o := ownable.New()

	std.TestSetRealm(std.NewUserRealm(secondCaller))
	result := NewFromOwnable(o)

	if result.Owner() != firstCaller {
		t.Fatalf("Expected %s, got %s\n", firstCaller, result.Owner())
	}
}

func TestSetUnpaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	result := New()
	result.Unpause()

	if result.IsPaused() {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.IsPaused())
	}
}

func TestSetPaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	result := New()
	result.Pause()

	if !result.IsPaused() {
		t.Fatalf("Expected result to be paused, got %t\n", result.IsPaused())
	}
}

func TestIsPaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	result := New()

	if result.IsPaused() {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.IsPaused())
	}

	result.Pause()

	if !result.IsPaused() {
		t.Fatalf("Expected result to be paused, got %t\n", result.IsPaused())
	}
}
