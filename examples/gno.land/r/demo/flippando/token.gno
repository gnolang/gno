package flippando

import (
	"errors"
	"strconv"
	"time"
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/flippandoserver"
	"gno.land/p/demo/avl"
)

type lockedVault struct {
	tokenId string
	lockedFlip uint64
}

var (
	flipToken = grc20.NewAdminToken("Flippando", "FLIP", 6)

	totalLockedSupply uint64 = 0
	totalSupply uint64 = 1000000000 // 1B per chain
	
	userFlipNFTs avl.Tree // onwer -> avl.Tree // []*lockedVault
)

func MintLocked(recipient std.Address, tokenId string, flipAmount uint64) error {
    flipToken.Mint(recipient, flipAmount)
    var userLockedBalances []*lockedVault
    
    v, exists := userFlipNFTs.Get(recipient.String())

    if exists {
        // Assert the type of v to []*lockedVault
        var ok bool
        userLockedBalances, ok = v.([]*lockedVault)
        if !ok {
            return errors.New("invalid type for user locked balances")
        }
    } 

    lockingBalance := &lockedVault{
        tokenId: tokenId,
        lockedFlip: flipAmount,
    }

    userLockedBalances = append([]*lockedVault{lockingBalance}, userLockedBalances...)
    userFlipNFTs.Set(recipient.String(), userLockedBalances)

    totalLockedSupply += flipAmount

    return nil
}


func UnlockAndTransferFLIP(recipient std.Address, tokenId string) error {
    var userLockedBalances []*lockedVault
    var flipAmount uint64
    tokenFound := false

    v, found := userFlipNFTs.Get(recipient.String())
    if found {
        // Assert the type of v to []*lockedVault
        var ok bool
        userLockedBalances, ok = v.([]*lockedVault)
        if !ok {
            return errors.New("invalid type for user locked balances")
        }

        // Create a new slice for storing balances without the unlocked token
        var newLockedBalances []*lockedVault
        for _, lv := range userLockedBalances {
            if lv.tokenId == tokenId {
                // Record the amount and indicate the token was found
                flipAmount = lv.lockedFlip
                tokenFound = true
            } else {
                // Add to the new slice if it's not the token to unlock
                newLockedBalances = append(newLockedBalances, lv)
            }
        }

        // Use the new slice if the token was found and removed
        if tokenFound {
            userLockedBalances = newLockedBalances
        }
    }

    if !tokenFound {
        return ErrInvalidTokenId // Ensure this error is defined
    }

    // Update the user's locked balances in userFlipNFTs
    userFlipNFTs.Set(recipient.String(), userLockedBalances)

    // Update the total locked supply
    totalLockedSupply -= flipAmount

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

// it is used in marketplace
func TransferFLIPToSeller(from, recipient std.Address, fPrice uint64){
	// seller approves the marketoplace contract owner to spend an amount equal with the total price
	// some of it will be burned later
	flipToken.Approve(from, std.CurrentRealm().Addr(), fPrice)
	totalFLIPBalance, err := flipToken.BalanceOf(from)
	if err != nil {
		panic("BalanceOf should not happen")
	}

	lockedFLIPBalance := getUserLockedBalances(from)

	// logic for burning a random amount from the fPrice
	if (totalFLIPBalance - lockedFLIPBalance >= fPrice) {	
		randomPercentage := flippandoserver.GenerateRandomNumbers(1, 1, 50, time.Now().Unix())
		fUpdatedPrice := fPrice - uint64(randomPercentage[0])
		flipToken.Transfer(from, recipient, fUpdatedPrice)
		flipToBurn := fPrice - fUpdatedPrice
		flipToken.Burn(from, flipToBurn)
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

func getUserLockedBalances(user std.Address) uint64 {
    lockedFLIPBalance := uint64(0)

    // Get the user's locked balances if they exist
    v, exists := userFlipNFTs.Get(user.String())
    if exists {
        // Assert the type of v to []*lockedVault
        userLockedBalances, ok := v.([]*lockedVault)
        if !ok {
            // Handle the type assertion failure as appropriate
            // For example, you could log an error or return 0
            return 0
        }

        // Iterate over the slice to sum up the locked balances
        for _, lv := range userLockedBalances {
            lockedFLIPBalance += lv.lockedFlip
        }
    }

    return lockedFLIPBalance
}

