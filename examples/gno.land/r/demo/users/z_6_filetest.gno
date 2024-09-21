package main

import (
	"std"

	"gno.land/r/demo/users"
)

const admin = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")

func main() {
	caller := std.GetOrigCaller()
	// as admin, grant invites to unregistered user.
	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller.String() + ":1")
	println("done")
}

// Error:
// invalid user g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
