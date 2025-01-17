package gnocket

import (
	"std"
	"strings"
	//"strconv"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/avl"
	"gno.land/r/matijamarjanovic/tokenhub"	
	//"gno.land/p/demo/ufmt"

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
		tokenBalances: avl.NewTree(),
	}

	wallets.Set(string(caller), wallet)

	braw := GetAllBalancesRaw()
	println(braw)


	return wallet
}

func GetAllTokens() string {
	return tokenhub.GetAllTokens()
}

func GetAllTokensNoNFTs() string {
	return tokenhub.GetAllTokensNoNFTs()
}


func Render(path string) string {


	sb := strings.Builder{}
	sb.WriteString("")

	return sb.String()
}
