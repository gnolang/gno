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

// GetUserTokenBalances returns a string of all the grc20 tokens the user owns
func GetUserTokenBalances(userNameOrAddress string) string {
	return getTokenBalances(userNameOrAddress, false)
}

// GetUserTokenBalancesNonZero returns a string of all the grc20 tokens the user owns, but only the ones that have a balance greater than 0
func GetUserTokenBalancesNonZero(userNameOrAddress string) string {
	return getTokenBalances(userNameOrAddress, true)
}

// GetUserNFTBalances returns a string of all the NFTs the user owns
func GetUserNFTBalances(userNameOrAddress string) string {
	return getNFTBalances(userNameOrAddress)
}

// GetUserMultiTokenBalances returns a string of all the multi-tokens the user owns
func GetUserMultiTokenBalances(userNameOrAddress string) string {
	return getMultiTokenBalances(userNameOrAddress, false)
}

// GetUserMultiTokenBalancesNonZero returns a string of all the multi-tokens the user owns, but only the ones that have a balance greater than 0
func GetUserMultiTokenBalancesNonZero(userNameOrAddress string) string {
	return getMultiTokenBalances(userNameOrAddress, true)
}

// GetToken returns a token instance for a given key
func GetToken(key string) *grc20.Token {
	return grc20reg.Get(key)
}

// MustGetToken returns a token instance for a given key, panics if the token is not found
func MustGetToken(key string) *grc20.Token {
	return grc20reg.MustGet(key)
}

// GetNFT returns an NFT instance for a given key
func GetNFT(key string) grc721.IGRC721 {
	nftGetter, ok := registeredNFTs.Get(key)
	if !ok {
		return nil
	}
	return (nftGetter.(grc721.NFTGetter))()
}

// MustGetNFT returns an NFT instance for a given key, panics if the NFT is not found
func MustGetNFT(key string) grc721.IGRC721 {
	nftGetter := GetNFT(key)
	if nftGetter == nil {
		panic("unknown NFT: " + key)
	}
	return nftGetter
}

// GetMultiToken returns a multi-token instance for a given key
func GetMultiToken(key string) grc1155.IGRC1155 {
	info, ok := registeredMTs.Get(key)
	if !ok {
		return nil
	}
	mt := info.(GRC1155TokenInfo).Collection
	return mt()
}

// MustGetMultiToken returns a multi-token instance for a given key, panics if the multi-token is not found
func MustGetMultiToken(key string) grc1155.IGRC1155 {
	info := GetMultiToken(key)
	if info == nil {
		panic("unknown multi-token: " + key)
	}
	return info
}

// GetAllNFTs returns a string of all the NFTs registered
func GetAllNFTs() string {
	var out string
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		out += ufmt.Sprintf("NFT:%s,", key)
		return false
	})
	return out
}

// GetAllTokens returns a string of all the tokens registered
func GetAllTokens() string {
	var out string
	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		out += "Token:" + key + ","
		return false
	})
	return out
}

// GetAllTokenWithDetails returns a string of all the tokens registered with their details
func GetAllTokenWithDetails() string {
	var out string
	grc20reg.GetRegistry().Iterate("", "", func(key string, value interface{}) bool {
		token := value.(*grc20.Token)
		out += ufmt.Sprintf("Token:%s,Name:%s,Symbol:%s,Decimals:%d;", key, token.GetName(), token.GetSymbol(), token.GetDecimals())
		return false
	})
	return out
}

// GetAllMultiTokens returns a string of all the multi-tokens registered
func GetAllMultiTokens() string {
	var out string
	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		out += "MultiToken:" + key + ","
		return false
	})
	return out
}

// GetAllRegistered returns a string of all the registered tokens, NFTs and multi-tokens
func GetAllRegistered() string {
	return GetAllNFTs() + GetAllTokens() + GetAllMultiTokens()
}

// getNFTBalances returns a string of all the NFTs the user owns
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

// getTokenBalances returns a string of all the tokens the user owns
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

// getMultiTokenBalances returns a string of all the multi-tokens the user owns
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

// getAddressForUsername returns an address for a given username or address
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
