// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main() {
	crossing()

	// Calling with a user that was not invited
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	commondao.New("Foo")
}

// Error:
// unauthorized
