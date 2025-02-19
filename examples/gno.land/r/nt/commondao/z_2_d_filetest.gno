// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	// Calling with a user that was not invited
	std.TestSetOrigCaller(user)
}

func main() {
	commondao.New("Foo")
}

// Error:
// unauthorized
