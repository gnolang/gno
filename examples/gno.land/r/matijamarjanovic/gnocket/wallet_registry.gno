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

// creates a new wallet for the caller but also updates the wallet balances
func NewWallet() *Wallet { 
	caller := std.PrevRealm().Addr()
	
	if wallet, exists := wallets.Get(string(caller)); exists {
		return wallet.(*Wallet)
	}

	wallet := &Wallet{
		Ownable: ownable.NewWithAddress(caller),
		tokenBalances: avl.NewTree(),
		public: true,
	}

	wallets.Set(string(caller), wallet)

	braw := GetAllBalancesRaw()
	braw = strings.TrimPrefix(braw, "Balances: ")

	allTokens := strings.Split(braw, ",")
	for _, token := range allTokens[:len(allTokens)-1] {
		tokenParts := strings.Split(token, ":")
		tokenType := tokenParts[0]
		tokenKey := tokenParts[1]
		tokenBalance := ""
		if tokenType != "NFT" {
			tokenBalance = tokenParts[2]
		}
		wallet.tokenBalances.Set(tokenKey, tokenBalance)
	}

	// TODO: research std.Emit (use it to listen for events and update the wallet balances)

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
