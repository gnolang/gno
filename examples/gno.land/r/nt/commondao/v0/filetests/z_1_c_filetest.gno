// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_1_c_filetest

package z_1_c_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	user    = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	invitee = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

func main(cur realm) {
	// Call as a users which is not a Common DAO member
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Invite(cross(cur), invitee)
}

// Error:
// unauthorized
