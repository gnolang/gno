package tokenhub

import (
	"strings"

	"gno.land/p/demo/tokens/grc1155"
	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/demo/defi/grc20reg"
	"gno.land/r/sys/users"
)

func GetUserTokenBalances(userNameOrAddress string) string {
	return getTokenBalances(userNameOrAddress, false)
}

func GetUserTokenBalancesNonZero(userNameOrAddress string) string {
	return getTokenBalances(userNameOrAddress, true)
}

func GetUserNFTBalances(userNameOrAddress string) string {
	return getNFTBalances(userNameOrAddress)
}

func GetUserMultiTokenBalances(userNameOrAddress string) string {
	return getMultiTokenBalances(userNameOrAddress, false)
}

func GetUserMultiTokenBalancesNonZero(userNameOrAddress string) string {
	return getMultiTokenBalances(userNameOrAddress, true)
}

func GetToken(key string) *grc20.Token {
	return grc20reg.Get(key)
}

func MustGetToken(key string) *grc20.Token {
	return grc20reg.MustGet(key)
}

// SECURITY: aggregator results are display data, not authority — never gate a
// transfer or mint on them; to mutate, call the owning realm's own wrapper.
func GetNFT(key string) *grc721.Token {
	v := registeredNFTs.Get(key)
	if v == nil {
		return nil
	}
	return v.(*grc721.Token)
}

func MustGetNFT(key string) *grc721.Token {
	nft := GetNFT(key)
	if nft == nil {
		panic("unknown NFT: " + key)
	}
	return nft
}

func GetMultiToken(key string) grc1155.IGRC1155 {
	info := registeredMTs.Get(key)
	if info == nil {
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
	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		out += "Token:" + key + ","
		return false
	})
	return out
}

func GetAllTokenWithDetails() string {
	var out string
	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		token := value.(*grc20.Token)
		out += ufmt.Sprintf("Token:%s,Name:%s,Symbol:%s,Decimals:%d;", key, token.GetName(), token.GetSymbol(), token.GetDecimals())
		return false
	})
	return out
}

func GetAllMultiTokens() string {
	var out string
	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		out += "MultiToken:" + key + ","
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
		nft := value.(*grc721.Token)
		key_parts := strings.Split(key, ".")
		owner, err := nft.OwnerOf(grc721.TokenID(key_parts[len(key_parts)-1]))
		if err == nil && addr == owner {
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

		token := value.(*grc20.Token)
		balance := token.BalanceOf(addr)
		if !nonZero || balance > 0 {
			out += ufmt.Sprintf("Token:%s:%d,", key, balance)
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
			if !nonZero || balance > 0 {
				out += ufmt.Sprintf("MultiToken:%s:%d,", key, balance)
			}
		}
		return false
	})

	return out
}

func getAddressForUsername(addrOrName string) address {
	addr := address(addrOrName)
	if addr.IsValid() {
		return addr
	}

	if userData, _ := users.ResolveName(addrOrName); userData != nil {
		return userData.Addr()
	}

	return ""
}
