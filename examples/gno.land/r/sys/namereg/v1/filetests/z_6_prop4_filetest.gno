// Test creating a GovDAO proposal to pause namereg realm using current pause value
package main

import (
	"gno.land/r/gov/dao"
	users "gno.land/r/sys/namereg/v1"
)

func main() {
	paused := users.IsPaused()
	dao.MustCreateProposal(cross, users.NewSetPausedExecutor(paused))
}

// Error:
// paused value is already false
