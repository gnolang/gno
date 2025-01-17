package gnocket

import (
	"std"
//	"strings"
	"gno.land/p/demo/ownable"
//	"gno.land/p/demo/avl"
	"gno.land/r/matijamarjanovic/tokenhub"
)

type Wallet struct {
	*ownable.Ownable
}

func assertOwnership() {
	caller := std.GetOrigCaller()
	w, exists := wallets.Get(string(caller))
	if !exists {
		panic("wallet not found")
	}
	w.(*Wallet).AssertCallerIsOwner()
}

func GetBalances() string {
	assertOwnership()
	return tokenhub.GetMyBalances()
}

func GetAllBalances() string {
	assertOwnership()

	return tokenhub.GetMyBalancesAll()
}

func Transfer(tokenPath string, to std.Address, amount uint64) string {
	assertOwnership()
	
	// Get the token getter from tokenhub
	tokenGetter := tokenhub.MustGetGRC20(tokenPath)
	token := tokenGetter()
	
	// Create an impersonation teller for the wallet owner
	//caller := std.GetOrigCaller()
	teller := token.CallerTeller()
	
	// Perform the transfer
	err := teller.Transfer(to, amount)
	if err != nil {
		panic(err)
	}

	return "Transfer successful"
}
