package flippando

import (
	"errors"
	"strconv"
	"strings"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/avl"
)

var (
	flipToken = grc20.NewAdminToken("Flippando", "FLIP", 6)

	totalLockedSupply uint64 = 0
	totalUnlockedSupply uint64 = 0 
	totalSupply uint64 = 1000000000 // 1B per chain
	
	userFlipNFTs avl.Tree // tokenId -> owner
	userLockedBalances avl.Tree // tokenId -> amount
)

func MintLocked(recipient std.Address, tokenId string, flipAmount uint64) error {
	flipToken.Mint(recipient, flipAmount)

	strFlipAmount := strconv.FormatUint(flipAmount, 10)

	userFlipNFTs.Set(tokenId, string(recipient))
	userLockedBalances.Set(tokenId, strFlipAmount)

	totalLockedSupply =+ flipAmount
	totalUnlockedSupply =- flipAmount

	return nil
}

func UnlockAndTransferFLIP(recipient std.Address, tokenId string) error {
	
	flipAmount, ferr := userLockedBalances.Get(tokenId)
	// add extra check by calling userFlipNFTs
	caller := std.PrevRealm().Addr()

	fAmount, err := strconv.Atoi(strings.TrimSpace(string(flipAmount)))

	flipToken.Transfer(caller, recipient, uint64(fAmount))

	totalLockedSupply =- uint64(fAmount)
	totalUnlockedSupply =+ uint64(fAmount)

	return nil
}

func GetFLIPBalance(user std.Address) uint64 {
	flipBalance, err := flipToken.BalanceOf(user)
	return flipBalance
}
