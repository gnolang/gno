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
	
	userFlipNFTs avl.Tree // tokenId -> owner
	userLockedBalances avl.Tree // tokenId -> amount
)

func MintLocked(recipient std.Address, tokenId string, flipAmount uint64) error {
	flipToken.Mint(recipient, flipAmount)

	strFlipAmount := strconv.FormatUint(flipAmount, 10)

	userFlipNFTs.Set(tokenId, string(recipient))
	userLockedBalances.Set(tokenId, string(strFlipAmount))

	totalLockedSupply =+ flipAmount 

	return nil
}

func UnlockAndTransferFLIP(recipient std.Address, tokenId string) error {
	
	flipAmount, found := userLockedBalances.Get(tokenId)
	if !found  {
		return "tokenId not found"
		//panic("UnlockAndTransferFlip tokenId not found")
	}
	// add extra check by calling userFlipNFTs
	flipOwner, exists := userFlipNFTs.Get(tokenId)
	if !exists {
		panic("tokenId doesn't have an owner")
	} else {
		if recipient.String() != flipOwner {
			panic("recipient is not owner")
		}
	}

	//caller := std.PrevRealm().Addr()
	caller :=std.CurrentRealm().Addr()

	fAmount, err := strconv.Atoi(strings.TrimSpace(string(flipAmount)))
	if err != nil {
		panic("UnlockAndTransferFlip strconv should not happen")
	}
	// replace this with deletion from userLockedBalances
	owner, tokenIdRemoved := userLockedBalances.Remove(tokenId)
	if !tokenIdRemoved {
		panic("token id not removed from userLockedBalances")
	}

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

	lockedFLIPBalance := uint64(0)
	userFlipNFTs.ReverseIterate("", "", func(tokenId string, owner interface{})bool {
		if(user.String() == string(owner)){
			lockedFlip, found := userLockedBalances.Get(tokenId)
			if !found {
				// it may mean it has been removed
			}
			lockedFlipAmount, err := strconv.Atoi(string(lockedFlip))
			lockedFLIPBalance =+ uint64(lockedFlipAmount)
		}
	})
	return lockedFLIPBalance
}
