package pausable

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
)

var (
	firstCaller  = std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	secondCaller = std.Address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

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
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	o := ownable.New()

	std.TestSetRealm(std.NewUserRealm(secondCaller))
	std.TestSetOrigCaller(secondCaller) // TODO(bug): should not be needed

	result := NewFromOwnable(o)

	if result.Owner() != firstCaller {
		t.Fatalf("Expected %s, got %s\n", firstCaller, result.Owner())
	}
}

func TestSetUnpaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()
	result.Unpause()

	if result.IsPaused() {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.IsPaused())
	}
}

func TestSetPaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()
	result.Pause()

	if !result.IsPaused() {
		t.Fatalf("Expected result to be paused, got %t\n", result.IsPaused())
	}
}

func TestIsPaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()

	if result.IsPaused() {
		t.Fatalf("Expected result to be unpaused, got %t\n", result.IsPaused())
	}

	result.Pause()

	if !result.IsPaused() {
		t.Fatalf("Expected result to be paused, got %t\n", result.IsPaused())
	}
}
