// PKGPATH: gno.land/r/sys/namereg/v1/filetests/z_6_prop4_filetest

// Test creating a GovDAO proposal to pause namereg realm using current pause value
package z_6_prop4_filetest

import (
	"gno.land/r/gov/dao"
	users "gno.land/r/sys/namereg/v1"
)

func main(cur realm) {
	paused := users.IsPaused()
	dao.MustCreateProposal(cross(cur), users.NewSetPausedExecutor(cross(cur), paused))
}

// Error:
// paused value is already false
