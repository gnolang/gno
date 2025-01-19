package gnocket

import (
	"errors"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/demo/users"
	"gno.land/r/matijamarjanovic/tokenhub"
)

var (
	wallets = avl.NewTree() // std.Address -> *Wallet
)

// creates a new wallet for the caller
func NewWallet() *Wallet {
	caller := std.PrevRealm().Addr()

	if wallet, exists := wallets.Get(string(caller)); exists {
		return wallet.(*Wallet)
	}

	wallet := &Wallet{
		Ownable:       ownable.NewWithAddress(caller),
		tokenBalances: avl.NewTree(),
		public:        true,
	}

	wallets.Set(string(caller), wallet)

	braw := GetAllBalancesRaw()
	braw = strings.TrimPrefix(braw, "Balances: ")


	//todo remove implementation of avltree for user balances
	allTokens := strings.Split(braw, ",")
	for _, token := range allTokens[:len(allTokens)-1] { 
		tokenParts := strings.Split(token, ":")
		tokenType := tokenParts[0]
		tokenKey := tokenParts[1]
		tokenBalance := ""
		userTM := &userTokenMetadata{}

		if tokenType != "NFT" {
			tokenBalance = tokenParts[2]
			balanceInt, err := strconv.Atoi(tokenBalance)

			if err != nil {
				continue
			}
			userTM = &userTokenMetadata{
				balance:   uint64(balanceInt),
				tokenType: tokenType,
			}
		} else {
			userTM = &userTokenMetadata{
				balance:   1,
				tokenType: tokenType,
			}
		}

		wallet.tokenBalances.Set(tokenKey, userTM)
	}

	// TODO: research std.Emit (use it to listen for events and update the wallet balances)

	return wallet
}

func GetMyBalances() string {
	caller := std.PrevRealm().Addr()
	wallet, exists := wallets.Get(string(caller))

	if !exists {
		return "no wallet found"
	}

	sb := strings.Builder{}

	wallet.(*Wallet).tokenBalances.Iterate("", "", func(key string, value interface{}) bool {
		sb.WriteString(value.(*userTokenMetadata).tokenType)
		sb.WriteString(":")
		sb.WriteString(key)
		if value.(*userTokenMetadata).tokenType != "NFT" {
			sb.WriteString(":")
			sb.WriteString(ufmt.Sprintf("%d", value.(*userTokenMetadata).balance))
		}
		sb.WriteString(",")
		return false
	})

	return sb.String()
}

func GetMyTokenBalances() string {
	caller := std.PrevRealm().Addr()
	wallet, exists := wallets.Get(string(caller))

	if !exists {
		return "no wallet found"
	}

	sb := strings.Builder{}

	wallet.(*Wallet).tokenBalances.Iterate("", "", func(key string, value interface{}) bool {
		if value.(*userTokenMetadata).tokenType == "Token" {
			sb.WriteString(value.(*userTokenMetadata).tokenType)
			sb.WriteString(":")
			sb.WriteString(key)
			sb.WriteString(":")
			sb.WriteString(ufmt.Sprintf("%d", value.(*userTokenMetadata).balance))
			sb.WriteString(",")
		}
		return false
	})

	if sb.Len() == 0 {
		return "no token found"
	}

	return sb.String()
}

func TogglePublic() error {
	caller := std.PrevRealm().Addr()
	wallet, exists := wallets.Get(string(caller))

	if !exists {
		return errors.New("no wallet found")
	}

	wallet.(*Wallet).public = !wallet.(*Wallet).public

	return nil
}

func GetAllTokens() string {
	return tokenhub.GetAllTokens()
}

func GetAllTokensNoNFTs() string {
	return tokenhub.GetAllTokensNoNFTs()
}

// TODO : add filtering by token type with parsing address.[tokenType]; if no token type is provided, return all tokens
func Render(path string) string {
	userToDisplay := getAddressForUsername(path)

	if RenderUserPageAllTypes(userToDisplay) != "" {
		return RenderUserPageAllTypes(userToDisplay)
	}

	sb := strings.Builder{}
	sb.WriteString("")

	return sb.String()
}

// TODO : add  sorting by token type
func RenderUserPageAllTypes(userToDisplay std.Address) string {

	wallet, exists := wallets.Get(string(userToDisplay))
	if !exists {
		return "no wallet found"
	}
	sb := strings.Builder{}

	sb.WriteString(md.H1(ufmt.Sprintf("Wallet of %s", userToDisplay)))

	if wallet.(*Wallet).public {
		sb.WriteString(md.BulletList([]string{}))

		wallet.(*Wallet).tokenBalances.Iterate("", "", func(key string, value interface{}) bool {
			tokenType := value.(*userTokenMetadata).tokenType
			tokenKey := key
			balance := ufmt.Sprintf("%d", value.(*userTokenMetadata).balance)
			if tokenType == "NFT" {
				sb.WriteString(md.BulletItem(ufmt.Sprintf("%s : %s", tokenType, tokenKey)))
			} else {
				sb.WriteString(md.BulletItem(ufmt.Sprintf("%s : %s : %s", tokenType, tokenKey, balance)))
			}
			return false
		})

	} else {
		return md.H1("Wallet is not public") + md.Paragraph("If this is your wallet, you can toggle it's visibility by triggering TogglePublic().")
	}

	return sb.String()
}

func getAddressForUsername(addrOrName string) std.Address {
	addr := std.Address(addrOrName)

	if addr.IsValid() {
		return addr
	}
	if user := users.GetUserByName(addrOrName); user != nil {
		return user.Address
	}

	return ""
}
