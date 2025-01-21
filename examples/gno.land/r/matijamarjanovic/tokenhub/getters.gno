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

func GetUserTokenBalances(userNameOrAddress string) string {
	return getTokenBalances(userNameOrAddress, false)
}

func GetUserNFTBalances(userNameOrAddress string) string {
	return getNFTBalances(userNameOrAddress)
}

func GetUserMultiTokenBalances(userNameOrAddress string) string {
	return getMultiTokenBalances(userNameOrAddress, false)
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

func getNFTBalances(input string) string {
	addr := getAddressForUsername(input)
	if !addr.IsValid() {
		panic("invalid address or username: " + input)
	}
	var out string

	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		key_parts := strings.Split(key, ".")
		owner, err := nft.OwnerOf(grc721.TokenID(key_parts[len(key_parts)-1]))
		if err == nil && addr == owner { // show only the nfts owner owns
			out += "NFT:" + key + ","
		}
		return false
	})

	return out
}

func getTokenBalances(input string, nonZero bool) string {
	addr := getAddressForUsername(input)
	if !addr.IsValid() {
		panic("invalid address or username: " + input)
	}
	var out string
	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		tokenGetter := value.(grc20.TokenGetter)
		token := tokenGetter()
		balance := token.BalanceOf(addr)
		if nonZero && balance > 0 {
			out += "Token:" + key + ":" + balance + ","
		} else if !nonZero {
			out += "Token:" + key + ":" + balance + ","
		}
		return false
	})

	return out
}

func getMultiTokenBalances(input string, nonZero bool) string {
	addr := getAddressForUsername(input)
	if !addr.IsValid() {
		panic("invalid address or username: " + input)
	}
	var out string

	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		info := value.(GRC1155TokenInfo)
		mt := info.Collection()
		balance, err := mt.BalanceOf(addr, grc1155.TokenID(info.TokenID))
		if err == nil {
			if nonZero && balance > 0 {
				out += "MultiToken:" + key + ":" + balance + ","
			} else if !nonZero {
				out += "MultiToken:" + key + ":" + balance + ","
			}
		}
		return false
	})

	return out
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
