// Package monit links a monitoring system with the chain in both directions.
//
// The agent will periodically call Incr() and verify that the value is always
// higher than the previously known one. The contract will store the last update
// time and use it to detect whether or not the monitoring agent is functioning
// correctly.
package monit

import (
	"chain/runtime"
	"time"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/watchdog/v0"
)

var (
	counter          int
	lastUpdate       time.Time
	lastCaller       address
	watchdogDuration = 5 * time.Minute
	wd               = watchdog.Watchdog{Duration: watchdogDuration}
	Ownable          = ownable.NewWithOrigin()
)

// Incr increments the counter and informs the watchdog that we're alive.
// This function can be called by anyone.
func Incr(cur realm) int {
	counter++
	lastUpdate = time.Now()
	lastCaller = runtime.PreviousRealm().Address()
	wd.Alive()
	return counter
}

// Reset resets the realm state.
// This function can only be called by the admin.
func Reset(cur realm) {
	Ownable.AssertOwnedByPrevious()
	counter = 0
	lastCaller = runtime.PreviousRealm().Address()
	lastUpdate = time.Now()
	wd = watchdog.Watchdog{Duration: watchdogDuration}
}

func Render(_ string) string {
	status := wd.Status()
	return ufmt.Sprintf(
		"counter=%d\nlast update=%s\nlast caller=%s\nstatus=%s",
		counter, lastUpdate, lastCaller, status,
	)
}
