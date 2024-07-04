// Package monit links a monitoring system with the chain in both directions.
//
// The agent will periodically call Incr() and verify that the value is always
// higher than the previously known one. The contract will store the last update
// time and use it to detect whether or not the monitoring agent is functioning
// correctly.
package monit

import (
	"std"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/watchdog"
)

var (
	counter          int
	wd               = watchdog.Watchdog{Duration: 5 * time.Minute}
	lastUpdate       time.Time
	lastCaller       std.Address
	owner            = ownable.New() // TODO: replace with -> ownable.NewWithAddress...
	watchdogDuration = 5 * time.Minute
)

// Incr increments the counter and informs the watchdog that we're alive.
// This function can be called by anyone.
func Incr() int {
	counter++
	lastUpdate = time.Now()
	lastCaller = std.PrevRealm().Addr()
	wd.Alive()
	return counter
}

// Reset resets the realm state.
// This function can only be called by the admin.
func Reset() {
	if owner.CallerIsOwner() != nil { // TODO: replace with owner.AssertCallerIsOwner
		panic("unauthorized")
	}
	counter = 0
	lastCaller = std.PrevRealm().Addr()
	lastUpdate = time.Now()
	wd = watchdog.Watchdog{Duration: 5 * time.Minute}
}

func Render(_ string) string {
	status := wd.Status()
	return ufmt.Sprintf(
		"counter=%d\nlast update=%s\nlast caller=%s\nstatus=%s",
		counter, lastUpdate, lastCaller, status,
	)
}

// TransferOwnership transfers ownership to a new owner. This is a proxy to
// ownable.Ownable.TransferOwnership.
func TransferOwnership(newOwner std.Address) { owner.TransferOwnership(newOwner) }
