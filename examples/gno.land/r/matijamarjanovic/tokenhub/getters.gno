package tokenhub

import (
	//"std"
	"strings"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/grc/grc1155"
	"gno.land/r/demo/grc20reg"
	"gno.land/p/demo/ufmt"
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
	var sb strings.Builder
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		sb.WriteString(ufmt.Sprintf("NFT:%s,", key))
		return false
	})
	return sb.String()
}

func GetAllTokens() string {
	var sb strings.Builder
	grc20reg.Iterate(func(key string, tokenGetter grc20.TokenGetter) bool {
		sb.WriteString(ufmt.Sprintf("Token:%s,", key))
		return false
	})
	return sb.String()
}

func GetAllMultiTokens() string {
	var sb strings.Builder
	registeredMTs.Iterate("", "", func(key string, value interface{}) bool {
		sb.WriteString(ufmt.Sprintf("MultiToken:%s,", key))
		return false
	})
	return sb.String()
}

func GetAllRegistered() string {
	return GetAllNFTs() + GetAllTokens() + GetAllMultiTokens()
}
