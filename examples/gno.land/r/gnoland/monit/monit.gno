// Package monit is intended to link a monitoring system with the chain in both directions.
//
// The agent will periodically call Incr() and verify that the value is always higher to the previously known one.
// The contract will store the last update time and will use it to detect wether or not the monitoring agent seems to be done or not.
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
	owner            = ownable.New() // TODO: replace with -> // owner          = ownable.NewWithAddress(std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")) // @manfred; TODO replace with the agent address
	watchdogDuration = 5 * time.Minute
)

// Anyone can call Incr.
func Incr() int {
	counter++
	lastUpdate = time.Now()
	lastCaller = std.PrevRealm().Addr()
	return counter
}

// Only the admin can Reset.
func Reset() {
	if owner.CallerIsOwner() != nil { // TODO: replace with owner.AsserCallerIsOwner
		panic("unauthorized")
	}
	counter = 0
	lastCaller = std.PrevRealm().Addr()
	lastUpdate = time.Now()
}

func Render(_ string) string {
	status := "KO"
	if time.Since(lastUpdate) < watchdogDuration {
		status = "OK"
	}
	return ufmt.Sprintf(
		"counter=%d\nlast update=%s\nlast caller=%s\nstatus=%s",
		counter, lastUpdate, lastCaller, status,
	)
}

// proxies to ownable.Ownable.*
func TransferOwnership(newOwner std.Address) { owner.TransferOwnership(newOwner) }
