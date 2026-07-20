// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_2_c_filetest

package z_2_c_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const user = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main(cur realm) {
	// Calling with a user that was not invited
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.New(cross(cur), "Foo")
}

// Error:
// unauthorized
