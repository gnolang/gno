// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/gnoland/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	// Calling with a user that was not invited
	std.TestSetOriginCaller(user)
}

func main() {
	commondao.New("Foo")
}

// Error:
// unauthorized
