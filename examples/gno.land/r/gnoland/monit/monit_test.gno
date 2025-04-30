package monit

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/watchdog"
)

func initTest() {
	counter = 0
	lastUpdate = time.Time{}
	lastCaller = std.Address("")
	wd = watchdog.Watchdog{Duration: watchdogDuration}
	creator := std.Address("g1creator")
	Ownable = ownable.NewWithAddress(creator)
}

func TestPackage(t *testing.T) {
	initTest()

	testing.SetRealm(std.NewUserRealm("g1user"))

	// initial state, watchdog is KO.
	{
		expected := `counter=0
last update=0001-01-01 00:00:00 +0000 UTC
last caller=
status=KO`
		got := Render("")
		uassert.Equal(t, expected, got)
	}

	// call Incr(), watchdog is OK.
	cross(Incr)()
	cross(Incr)()
	cross(Incr)()
	{
		expected := `counter=3
last update=2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
last caller=g1user
status=OK`
		got := Render("")
		uassert.Equal(t, expected, got)
	}

	/* XXX: improve tests once we've the missing std.TestSkipTime feature
		// wait 1h, watchdog is KO.
		use std.TestSkipTime(time.Hour)
		{
			expected := `counter=3
	last update=2009-02-13 22:31:30 +0000 UTC m=+1234564290.000000001
	last caller=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	status=KO`
			got := Render("")
			uassert.Equal(t, expected, got)
		}

		// call Incr(), watchdog is OK.
		Incr()
		{
			expected := `counter=4
	last update=2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
	last caller=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	status=OK`
			got := Render("")
			uassert.Equal(t, expected, got)
		}
	*/
}

func TestReset(t *testing.T) {
	initTest()

	// Initial state check
	initialCounter := counter
	initialLastUpdate := lastUpdate
	initialLastCaller := lastCaller
	initialStatus := wd.Status()

	// Call Incr to change the state
	user := std.Address("g1user")
	testing.SetRealm(std.NewUserRealm(user))
	cross(Incr)()
	uassert.True(t, counter > initialCounter, "counter should have increased after Incr")
	uassert.True(t, lastUpdate.After(initialLastUpdate), "lastUpdate should have been updated after Incr")
	uassert.Equal(t, user, lastCaller, "lastCaller mismatch")
	uassert.NotEqual(t, initialStatus, wd.Status(), "watchdog status should have changed after Incr") // Status changes after Alive() is called

	// Call Reset as the owner
	ownerAddr := Ownable.Owner()
	testing.SetRealm(std.NewUserRealm(ownerAddr)) // Simulate call from the owner
	cross(Reset)()
	uassert.Equal(t, 0, counter, "counter should be 0 after Reset")
	uassert.Equal(t, ownerAddr, lastCaller, "lastCaller should be the owner address after Reset")
	uassert.Equal(t, watchdogDuration.String(), wd.Duration.String(), "watchdog duration mismatch after Reset")
	uassert.Equal(t, "KO", wd.Status(), "watchdog status should be KO after Reset")
}
