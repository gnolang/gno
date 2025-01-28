package pausable

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/urequire"
)

var (
	firstCaller = std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	o           = ownable.NewWithAddress(firstCaller)
)

func TestNewFromOwnable(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)

	result := NewFromOwnable(o)
	urequire.Equal(t, firstCaller.String(), result.Owner().String())
}

func TestSetUnpaused(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)
	result := NewFromOwnable(o)

	result.Unpause()
	urequire.False(t, result.IsPaused(), "Expected result to be unpaused")
}

func TestSetPaused(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)
	result := NewFromOwnable(o)

	result.Pause()
	urequire.True(t, result.IsPaused(), "Expected result to be paused")
}

func TestIsPaused(t *testing.T) {
	result := NewFromOwnable(o)
	urequire.False(t, result.IsPaused(), "Expected result to be unpaused")

	std.TestSetOrigCaller(firstCaller)
	result.Pause()
	urequire.True(t, result.IsPaused(), "Expected result to be paused")
}
