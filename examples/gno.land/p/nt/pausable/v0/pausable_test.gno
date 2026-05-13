package pausable

import (
	"testing"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	firstCaller = address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	o           = ownable.NewWithAddress(firstCaller)
)

func TestNewFromOwnable(t *testing.T) {
	result := NewFromOwnable(o)

	urequire.Equal(t, firstCaller.String(), result.Ownable().Owner().String())
}

func TestSetUnpaused(t *testing.T) {
	result := NewFromOwnable(o)

	testing.SetRealm(testing.NewUserRealm(firstCaller))
	func(cur realm) {
		urequire.NoError(t, result.Unpause(0, cur))
	}(cross)
	uassert.False(t, result.IsPaused(), "Expected result to be unpaused")
}

func TestSetPaused(t *testing.T) {
	result := NewFromOwnable(o)

	testing.SetRealm(testing.NewUserRealm(firstCaller))
	func(cur realm) {
		urequire.NoError(t, result.Pause(0, cur))
	}(cross)
	uassert.True(t, result.IsPaused(), "Expected result to be paused")
}

func TestIsPaused(t *testing.T) {
	result := NewFromOwnable(o)
	urequire.False(t, result.IsPaused(), "Expected result to be unpaused")

	testing.SetRealm(testing.NewUserRealm(firstCaller))
	func(cur realm) {
		urequire.NoError(t, result.Pause(0, cur))
	}(cross)
	uassert.True(t, result.IsPaused(), "Expected result to be paused")
}

func TestOwnable(t *testing.T) {
	result := NewFromOwnable(o)

	uassert.Equal(t, result.Ownable().Owner(), o.Owner())
}
