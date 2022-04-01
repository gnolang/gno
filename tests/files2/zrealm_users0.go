// PKGPATH: gno.land/r/users_test
// SEND: 2000gnot
package users_test

import (
	"gno.land/r/users"
)

func main() {
	// caller := std.GetOrigCaller()
	// addr1 := testutils.TestAddress("addr1")
	err := users.Register("", "jaekwon", "my profile")
	if err != nil {
		println(err.Error())
		panic(err)
	}
	println("done")
}

// Error:
// insufficient payment
