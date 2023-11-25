package flippando

import (
	"errors"
	"strconv"
	"strings"
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/avl"
)

var (
	flipToken = grc20.NewAdminToken("Flippando", "FLIP", 6)

	totalLockedSupply uint64 = 0
	totalSupply uint64 = 1000000000 // 1B per chain
	
	userFlipNFTs avl.Tree // onwer -> avl.Tree // tokenId -> flipAmount
)

func MintLocked(recipient std.Address, tokenId string, flipAmount uint64) error {

	//var userLockedBalances avl.Tree

	flipToken.Mint(recipient, flipAmount)

	strFlipAmount := strconv.FormatUint(flipAmount, 10)

	v, exists := userFlipNFTs.Get(recipient.String())
	if exists {
		userLockedBalances := v.(avl.Tree)
		userLockedBalances.Set(tokenId, strFlipAmount)
		userFlipNFTs.Set(recipient.String(), userLockedBalances)
	} else {
		firstBalance := avl.Tree{}
		firstBalance.Set(tokenId, strFlipAmount)
		userFlipNFTs.Set(recipient.String(), firstBalance)
	}

	totalLockedSupply =+ flipAmount 

	return nil
}

func UnlockAndTransferFLIP(recipient std.Address, tokenId string) error {
	var userLockedBalances avl.Tree
	v, found := userFlipNFTs.Get(recipient.String())
	if found  {
		userLockedBalances = v.(avl.Tree)
	} else {
		return ErrUnlockingRecipientNotFound
	}
	// check if the token exists in user locked balances
	flipToUnlock, exists := userLockedBalances.Get(tokenId)
	if !exists {
		return ErrInvalidTokenId
	} 

	fAmount, err := strconv.Atoi(strings.TrimSpace(string(flipToUnlock)))
	if err != nil {
		panic("UnlockAndTransferFlip strconv should not happen")
	}
	// replace this with deletion from userLockedBalances
	newLockedBalances, tokenIdRemoved := userLockedBalances.Remove(tokenId)
	if !tokenIdRemoved {
		panic("token id not removed from userLockedBalances")
	}
	
	userFlipNFTs.Set(recipient.String(), newLockedBalances)

	totalLockedSupply =- uint64(fAmount)

	return nil
}

// add a new Tranfer function, which will call the grc20.Transfer
// only if the user has available tokens to spend

func Transfer(from, recipient std.Address, fAmount uint64){
	// who's calling? std.PrevRealm() = msg.sender?

	// check if we can bypass this function by calling the interface directly, I don't think so
	
	totalFLIPBalance, err := flipToken.BalanceOf(from)
	if err != nil {
		panic("BalanceOf should not happen")
	}

	lockedFLIPBalance := getUserLockedBalances(from)

	if (totalFLIPBalance - lockedFLIPBalance >= fAmount) {
		flipToken.Transfer(from, recipient, fAmount)
	} else {
		panic("not enough available balance")
	}
}

func GetFLIPBalance(user std.Address) string {
	totalFLIPBalance, err := flipToken.BalanceOf(user)
	if err != nil {
		panic("BalanceOf should not happen")
	}

	lockedFLIPBalance := getUserLockedBalances(user)
	availableFLIPBalance := totalFLIPBalance - lockedFLIPBalance

	userBalances := &userFLIPBalances{
		lockedBalance: strconv.FormatUint(lockedFLIPBalance, 10),
		availableBalance: strconv.FormatUint(availableFLIPBalance, 10),
	}

	userBalancesJSON, err := userBalances.MarshalJSON() 
	if(err != nil){
		panic("error in marshalling user FLIP balances")
	}

	return string(userBalancesJSON)
}

func getUserLockedBalances (user std.Address) uint64 {
	var userLockedBalances avl.Tree
	lockedFLIPBalance := uint64(0)
	v, exists := userFlipNFTs.Get(user.String())
	if exists {
		userLockedBalances = v.(avl.Tree)
		userLockedBalances.ReverseIterate("", "", func(tokenId string, lockedFlip interface{})bool {
			lockedFlipAmount, err := strconv.Atoi(string(lockedFlip))
			lockedFLIPBalance =+ uint64(lockedFlipAmount)
		
		})	
	}

	return lockedFLIPBalance
}
