package tokenhub

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/r/demo/grc20reg"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"

)

func GetUserBalancesWithZeros(userNameOrAddress string) string {
	return getBalances(userNameOrAddress)
}

func GetUserBalances(userNameOrAddress string) string {
	return getNonZeroBalances(userNameOrAddress)
}

func GetToken(key string) *grc20.Token {
	getter := grc20reg.Get(key)
	return getter()
}

func MustGetToken(key string) *grc20.Token {
	getter := grc20reg.MustGet(key)
	return getter()
}

func GetNFT(key string) grc721.IGRC721 {
	nftGetter, ok := registeredNFTs.Get(key)
	if !ok {
		return nil
	}
	nft := nftGetter.(grc721.IGRC721)
	return nft
}

func MustGetNFT(key string) grc721.IGRC721 {
	nftGetter := GetNFT(key)
	if nftGetter == nil {
		panic("unknown NFT: " + key)
	}
	return nftGetter
}

func GetMultiToken(key string) grc1155.IGRC1155 {
	info, ok := registeredMTs.Get(key)
	if !ok {
		return nil
	}
	mt := info.(GRC1155TokenInfo).Collection
	return mt()
}

func MustGetMultiToken(key string) grc1155.IGRC1155 {
	info := GetMultiToken(key)
	if info == nil {
		panic("unknown multi-token: " + key)
	}
	return info
}

func GetAllNFTs() string {
	var out string
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		out += ufmt.Sprintf("NFT:%s,", key)
		return false
	})
	return out
}

func GetAllTokens() string {
	var out string
	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool{
		out += ufmt.Sprintf("Token:%s,", key)
		return false
	})
	return out
}

func GetAllMultiTokens() string {
	var out string
	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		out += ufmt.Sprintf("MultiToken:%s,", key)
		return false
	})
	return out
}

func GetAllRegistered() string {
	return GetAllNFTs() + GetAllTokens() + GetAllMultiTokens()
}


func getBalances(input string) string {
	addr := getAddressForUsername(input)
	if addr == "" {
		panic("invalid address or username: " + input)
	}

	balances := avl.NewTree()
	var sb strings.Builder

	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		key_parts := strings.Split(key, ".")
		owner, err := nft.OwnerOf(grc721.TokenID(key_parts[len(key_parts)-1]))
		if err == nil && addr == owner { // show only the nfts owner owns
			balances.Set("nft:"+key, 1)
		}
		return false
	})

	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		tokenGetter := value.(grc20.TokenGetter)
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		balances.Set("token:"+key, balance)
		return false
	})

	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mt := info.Collection()
		balance, err := mt.BalanceOf(addr, grc1155.TokenID(info.TokenID))
		if err == nil {
			balances.Set("mt:"+key, balance)
		}
		return false
	})

	sb.WriteString("Balances: ")
	balances.Iterate("", "", func(key string, value interface{}) bool {
		switch {
		case strings.HasPrefix(key, "nft:"):
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT:%s,", symbol))
		case strings.HasPrefix(key, "token:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "token:")
			sb.WriteString(ufmt.Sprintf("Token:%s:%d,", symbol, balance))
		case strings.HasPrefix(key, "mt:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "mt:")
			sb.WriteString(ufmt.Sprintf("MultiToken:%s:%d,",
				symbol,
				balance))
		}
		return false
	})

	return sb.String()
}

func getNonZeroBalances(input string) string {
	addr := getAddressForUsername(input)
	if addr == "" {
		panic("invalid address or username: " + input)
	}

	balances := avl.NewTree()
	var sb strings.Builder

	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		key_parts := strings.Split(key, ".")
		owner, err := nft.OwnerOf(grc721.TokenID(key_parts[len(key_parts)-1]))
		if err == nil && addr == owner { // show only the nfts owner owns
			balances.Set("nft:"+key, 1)
		}
		return false
	})

	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		tokenGetter := value.(grc20.TokenGetter)
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		if balance > 0 {
			balances.Set("token:"+key, balance)
		}
		return false
	})

	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mt := info.Collection()
		balance, err := mt.BalanceOf(addr, grc1155.TokenID(info.TokenID))
		if err == nil && balance > 0 {
			balances.Set("mt:"+key, balance)
		}
		return false
	})

	sb.WriteString("Non-Zero Balances: ")
	balances.Iterate("", "", func(key string, value interface{}) bool {
		switch {
		case strings.HasPrefix(key, "nft:"):
			symbol := strings.TrimPrefix(key, "nft:")
			sb.WriteString(ufmt.Sprintf("NFT:%s,", symbol))
		case strings.HasPrefix(key, "token:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "token:")
			sb.WriteString(ufmt.Sprintf("Token:%s:%d,", symbol, balance))
		case strings.HasPrefix(key, "mt:"):
			balance := value.(uint64)
			symbol := strings.TrimPrefix(key, "mt:")
			sb.WriteString(ufmt.Sprintf("MultiToken:%s:%d,",
				symbol,
				balance))
		}
		return false
	})

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
