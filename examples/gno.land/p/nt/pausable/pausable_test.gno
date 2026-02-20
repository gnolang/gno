package pausable

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/ownable"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

var (
	firstCaller = address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	o           = ownable.NewWithAddress(firstCaller)
)

func TestNewFromOwnable(t *testing.T) {
	testing.SetOriginCaller(firstCaller)
	result := NewFromOwnable(o)

	urequire.Equal(t, firstCaller.String(), result.Ownable().Owner().String())
}

func TestSetUnpaused(t *testing.T) {
	testing.SetOriginCaller(firstCaller)
	result := NewFromOwnable(o)

	result.Unpause()
	uassert.False(t, result.IsPaused(), "Expected result to be unpaused")
}

func TestSetPaused(t *testing.T) {
	testing.SetOriginCaller(firstCaller)
	result := NewFromOwnable(o)

	result.Pause()
	uassert.True(t, result.IsPaused(), "Expected result to be paused")
}

func TestIsPaused(t *testing.T) {
	result := NewFromOwnable(o)
	urequire.False(t, result.IsPaused(), "Expected result to be unpaused")

	testing.SetOriginCaller(firstCaller)
	result.Pause()
	uassert.True(t, result.IsPaused(), "Expected result to be paused")
}

func TestOwnable(t *testing.T) {
	result := NewFromOwnable(o)

	uassert.Equal(t, result.Ownable().Owner(), o.Owner())
}

func TestPauseRespectsPreviousMode(t *testing.T) {
	prevOwned := ownable.NewWithAddressByPrevious(firstCaller)
	result := NewFromOwnable(prevOwned)

	testing.SetRealm(testing.NewUserRealm(firstCaller))
	crossThrough(testing.NewCodeRealm("gno.land/r/test/test"), func() {
		urequire.NoError(t, result.Pause())
		uassert.True(t, result.IsPaused(), "Expected result to be paused")

		urequire.NoError(t, result.Unpause())
		uassert.False(t, result.IsPaused(), "Expected result to be unpaused")
	})
}

func crossThrough(rlm runtime.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}
