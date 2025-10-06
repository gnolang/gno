package monit

import (
	"testing"
	"time"

	"gno.land/p/nt/ownable"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/watchdog"
)

func initTest() {
	counter = 0
	lastUpdate = time.Time{}
	lastCaller = address("")
	wd = watchdog.Watchdog{Duration: watchdogDuration}
	creator := address("g1creator")
	Ownable = ownable.NewWithAddress(creator)
}

func TestPackage(t *testing.T) {
	initTest()

	testing.SetRealm(testing.NewUserRealm("g1user"))

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
	Incr(cross)
	Incr(cross)
	Incr(cross)
	{
		expected := `counter=3
last update=2009-02-13 23:31:30 +0000 UTC
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
	last update=2009-02-13 22:31:30 +0000 UTC
	last caller=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
	status=KO`
			got := Render("")
			uassert.Equal(t, expected, got)
		}

		// call Incr(), watchdog is OK.
		Incr()
		{
			expected := `counter=4
	last update=2009-02-13 23:31:30 +0000 UTC
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
	initialStatus := wd.Status()

	// Call Incr to change the state
	user := address("g1user")
	testing.SetRealm(testing.NewUserRealm(user))
	Incr(cross)
	uassert.True(t, counter > initialCounter, "counter should have increased after Incr")
	uassert.True(t, lastUpdate.After(initialLastUpdate), "lastUpdate should have been updated after Incr")
	uassert.Equal(t, user, lastCaller, "lastCaller mismatch")
	uassert.NotEqual(t, initialStatus, wd.Status(), "watchdog status should have changed after Incr") // Status changes after Alive() is called

	// Call Reset as the owner
	ownerAddr := Ownable.Owner()
	testing.SetRealm(testing.NewUserRealm(ownerAddr)) // Simulate call from the owner
	Reset(cross)
	uassert.Equal(t, 0, counter, "counter should be 0 after Reset")
	uassert.Equal(t, ownerAddr, lastCaller, "lastCaller should be the owner address after Reset")
	uassert.Equal(t, watchdogDuration.String(), wd.Duration.String(), "watchdog duration mismatch after Reset")
	uassert.Equal(t, "KO", wd.Status(), "watchdog status should be KO after Reset")
}
