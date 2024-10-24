package pausable

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/urequire"
)

var (
	firstCaller  = std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	secondCaller = std.Address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()

	urequire.False(t, result.paused, "Expected result to be unpaused")
	urequire.Equal(t, firstCaller.String(), result.Owner().String())
}

func TestNewFromOwnable(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	o := ownable.New()

	std.TestSetRealm(std.NewUserRealm(secondCaller))
	std.TestSetOrigCaller(secondCaller) // TODO(bug): should not be needed

	result := NewFromOwnable(o)

	urequire.Equal(t, firstCaller.String(), result.Owner().String())
}

func TestSetUnpaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()
	result.Unpause()

	urequire.False(t, result.IsPaused(), "Expected result to be unpaused")
}

func TestSetPaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()
	result.Pause()

	urequire.True(t, result.IsPaused(), "Expected result to be paused")
}

func TestIsPaused(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	result := New()
	urequire.False(t, result.IsPaused(), "Expected result to be unpaused")

	result.Pause()
	urequire.True(t, result.IsPaused(), "Expected result to be paused")
}
