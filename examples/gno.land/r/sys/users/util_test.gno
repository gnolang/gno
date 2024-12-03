package users

import "std"

var (
	whitelistedCallerAddr = std.DerivePkgAddr(gnolandUsers)
	adminRealm            = std.NewUserRealm(adminAddr)
	alice                 = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	aliceCodeRealm        = std.NewCodeRealm("gno.land/r/alice/package")
	aliceUsername         = "alice123"
)
