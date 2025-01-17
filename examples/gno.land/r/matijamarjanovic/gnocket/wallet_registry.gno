package gnocket

import (
	"std"
	"strings"
	"strconv"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/avl"
	"gno.land/r/matijamarjanovic/tokenhub"	
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/grc/grc721"
)

var (
	wallets = avl.NewTree()        // std.Address -> *Wallet
)

func NewWallet() *Wallet {
	caller := std.PrevRealm().Addr()
	
	if wallet, exists := wallets.Get(string(caller)); exists {
		return wallet.(*Wallet)
	}

	wallet := &Wallet{
		Ownable: ownable.NewWithAddress(caller),
	}
	
	wallets.Set(string(caller), wallet)
	return wallet
}

func GetAllTokens() string {
	return tokenhub.GetAllTokens()
}

func GetBalancesForRender() string {
	return tokenhub.GetBalancesAll("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
}

func Render(path string) string {


	sb := strings.Builder{}
	sb.WriteString(GetBalancesForRender())

	return sb.String()
}
