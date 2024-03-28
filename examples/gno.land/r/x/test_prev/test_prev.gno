package test_prev

import (
	"std"

	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/foo20"
)

func DoSomeWithUserBalance() string {
	caller := std.GetOrigCaller()
	balance := foo20.BalanceOf(pusers.AddressOrName(caller))

	if balance > 100 {
		return "rich user"
	}
	return "poor user"
}
