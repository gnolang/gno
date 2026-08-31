// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_18_b_filetest

package z_18_b_filetest

import (
	"strings"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Purpose is capped at 250 characters.
	commondao.New(cross(cur), "Foo", strings.Repeat("A", 251), "", "")
}

// Error:
// DAO purpose is too long, max length is 250 characters
