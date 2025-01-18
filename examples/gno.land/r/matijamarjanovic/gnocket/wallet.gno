package gnocket

import (
	"std"
//	"strings"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/avl"
	"gno.land/r/matijamarjanovic/tokenhub"
)

type Wallet struct {
	*ownable.Ownable
	tokenBalances *avl.Tree // string -> uint64
}

func assertOwnership() {
	caller := std.GetOrigCaller()
	w, exists := wallets.Get(string(caller))
	if !exists {
		panic("wallet not found")
	}
	w.(*Wallet).AssertCallerIsOwner()
}

func GetBalancesRaw() string {
	assertOwnership()
	return tokenhub.GetMyBalancesOrigCaller()
}

func GetAllBalancesRaw() string {
	assertOwnership()
	return tokenhub.GetMyBalancesAllOrigCaller()
}


