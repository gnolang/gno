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

	balances = avl.NewTree()       // std.Address -> avl.Tree(string -> uint64) : userAddress -> tokenBalances(tokenKey -> balance)
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
	useWallet()
	return wallet
}

// useWallet is a function that transfers all token funds from the user address
// to the wallet address in order to let the wallet be the real holder of funds
// Note: this means wallet should keep track of how much which user has, also 
// there should be a way to check balances of user in his wallet
func useWallet() {
	caller := std.PrevRealm().Addr()
	balances.Set(string(caller), avl.NewTree()) 
	callers_tree_uncast, _ := balances.Get(string(caller))
	callers_tree := callers_tree_uncast.(*avl.Tree)

	get_bal_result := tokenhub.GetBalancesAll(caller.String())
	only_balances := strings.Split(get_bal_result, " ")[1]

	tokens := strings.Split(only_balances, ",")

	for _, token := range tokens[:len(tokens)-1] {
		token_parts := strings.Split(token, ":")
		token_identifier := ufmt.Sprintf("%s:%s", token_parts[0], token_parts[1])
		token_balance := token_parts[2]

		callers_tree.Set(token_identifier, token_balance)
	}

	callers_tree.Iterate("", "", func(key string, value interface{}) bool {
		token_identifier_parts := strings.Split(key, ":")
		token_type := token_identifier_parts[0]
		token_key := token_identifier_parts[1]
		token_balance, _ := strconv.Atoi(value.(string))

		if token_type == "MultiToken" {
			key_parts := strings.Split(token_key, ".")
			token_id := key_parts[len(key_parts)-1]
			tokenInfo := tokenhub.MustGetGRC1155(token_key)
			tokenGetter := tokenInfo.Collection
			grc1155_token := tokenGetter()

			grc1155_token.SafeTransferFrom(caller, std.CurrentRealm().Addr(), grc1155.TokenID(token_id), uint64(token_balance))
			println("token_id", token_id, "token_balance", token_balance)

		} else if token_type == "NFT" {
			key_parts := strings.Split(token_key, ".")
			token_id := key_parts[len(key_parts)-1]
			tokenGetter := tokenhub.MustGetGRC721(token_key)
			nft_token := tokenGetter()
			nft_token.SafeTransferFrom(caller, std.CurrentRealm().Addr(), grc721.TokenID(token_id))
			println("token_id", token_id)

		}
		return true
	})

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
