package flippando

import (
	"errors"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/flippandoserver"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

type lockedVault struct {
	tokenId   string
	lockedFlip std.Coin
}

var (
	flipToken          = grc20.NewBanker("Flippando", "FLIP", 12)
    // 1000 uFLIP = 1 FLIP
    // max denom 1,000,000,000 FLIP
	totalLockedSupply  std.Coin = std.NewCoin("FLIP", 0)
	// hardcoding the airdrop account and restricting minting to it
	airdropAccount std.Address = std.Address("g1mphljw4duhzwqs72hv9grm4k6ufkvts3g7t3t4")
	userFlipNFTs       avl.Tree // onwer -> avl.Tree // []*lockedVault
)

// overrides the Mint function in the interface, restricts only to airdropAccount calls
func Mint(recipient std.Address, flipAmount std.Coin) error {
	if airdropAccount.String() != std.GetCaller().String() {
		return errors.New("invalid account calling Mint")
	}

	if airdropAccount.String() != recipient.String() {
		return errors.New("invalid recipient for Mint")
	}
	
	if err := flipToken.Mint(recipient, uint64(flipAmount.Amount)); err != nil {
		return errors.New(ufmt.Sprintf("Error minting FLIP to airdrop account: %s", err.Error()))
	}
}

// mint tokens and locks them, making them unavaialble for spending
func MintAndLockFLIP(recipient std.Address, tokenId string, flipAmount std.Coin) error {
	if err := flipToken.Mint(recipient, uint64(flipAmount.Amount)); err != nil {
		return err
	}

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
		tokenId:   tokenId,
		lockedFlip: flipAmount,
	}

	userLockedBalances = append([]*lockedVault{lockingBalance}, userLockedBalances...)
	userFlipNFTs.Set(recipient.String(), userLockedBalances)

	totalLockedSupply = totalLockedSupply.Add(flipAmount)

	return nil
}

// unlocks the balance, making it avaialble for spending for recipient
// no actual transfer is performed
func ReleaseLockedFLIP(recipient std.Address, tokenId string) error {
	var userLockedBalances []*lockedVault
	var flipAmount std.Coin
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
	totalLockedSupply = totalLockedSupply.Sub(flipAmount)

	return nil
}

// transfers fungible tokens, taking into account the locked balances
// i.e. only avaialble tokens can be spent
func Transfer(from, recipient std.Address, fAmount std.Coin) error {
	totalFLIPBalance := flipToken.BalanceOf(from)
	
	lockedFLIPBalance := getUserLockedBalances(from)

	if int64(totalFLIPBalance) >= fAmount.Amount && int64(totalFLIPBalance)-int64(lockedFLIPBalance.Amount) >= fAmount.Amount {
		if err := flipToken.Transfer(from, recipient, uint64(fAmount.Amount)); err != nil {
			return err
		}
	} else {
		return errors.New("not enough available balance")
	}

	return nil
}

// pays the FLIP price in a marketplace transaction
// burning a percentage between 1% anf 50%
func PayFLIPToSeller(from, recipient std.Address, fPrice std.Coin) (flipPaid string, flipBurned string, err error) {
	flipToken.Approve(from, std.CurrentRealm().Addr(), uint64(fPrice.Amount))
	totalFLIPBalance := flipToken.BalanceOf(from)
	
	lockedFLIPBalance := getUserLockedBalances(from)

	if int64(totalFLIPBalance) >= fPrice.Amount && int64(totalFLIPBalance)-lockedFLIPBalance.Amount >= fPrice.Amount {
		randomPercentage := flippandoserver.GenerateRandomNumbers(2, 1, 50, time.Now().Unix())
		randomBurnAmount := fPrice.Amount * int64(randomPercentage[0]) / 100
		fPaidPrice := fPrice.Amount - randomBurnAmount
		flipToBurn := fPrice.Amount - fPaidPrice 

		if err := flipToken.Transfer(from, recipient, uint64(fPaidPrice)); err != nil {
			return "", "", err
		}
		if err := flipToken.Burn(from, uint64(flipToBurn)); err != nil {
			return "", "", err
		}

		return strconv.FormatInt(fPaidPrice, 10), strconv.FormatInt(flipToBurn, 10), nil
	} else {
		return "", "", ErrFlipTokenNotEnoughBalance
	}
}

// transfers the locked tokens to a new recipient (toBuyer)
// used when a basic NFT is sold in the marketplace
func TransferLockedFLIP(fromSeller, toBuyer std.Address, tokenId string) error {
    var sellerLockedBalances []*lockedVault
    var flipAmount std.Coin
    tokenFound := false

    // Step 1: Unlock the tokens associated with the tokenId for the seller
    v, found := userFlipNFTs.Get(fromSeller.String())
    if found {
        // Assert the type of v to []*lockedVault
        var ok bool
        sellerLockedBalances, ok = v.([]*lockedVault)
        if !ok {
            return errors.New("invalid type for user locked balances")
        }

        // Create a new slice for storing balances without the unlocked token
        var newLockedBalances []*lockedVault
        for _, lv := range sellerLockedBalances {
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
            sellerLockedBalances = newLockedBalances
        }
    }

    if !tokenFound {
        return ErrInvalidTokenId // Ensure this error is defined
    }

    // Update the seller's locked balances in userFlipNFTs
    userFlipNFTs.Set(fromSeller.String(), sellerLockedBalances)

    // Update the total locked supply
    totalLockedSupply = totalLockedSupply.Sub(flipAmount)

    // Step 2: Transfer the unlocked tokens to the buyer
    if err := flipToken.Transfer(fromSeller, toBuyer, uint64(flipAmount.Amount)); err != nil {
        return err
    }

    // Step 3: Lock the received tokens in the buyer's userFlipNFTs
    var buyerLockedBalances []*lockedVault

    v, exists := userFlipNFTs.Get(toBuyer.String())
    if exists {
        // Assert the type of v to []*lockedVault
        var ok bool
        buyerLockedBalances, ok = v.([]*lockedVault)
        if !ok {
            return errors.New("invalid type for user locked balances")
        }
    }

    lockingBalance := &lockedVault{
        tokenId:   tokenId,
        lockedFlip: flipAmount,
    }

    buyerLockedBalances = append([]*lockedVault{lockingBalance}, buyerLockedBalances...)
    userFlipNFTs.Set(toBuyer.String(), buyerLockedBalances)

    totalLockedSupply = totalLockedSupply.Add(flipAmount)

    return nil
}

func GetFLIPBalance(user std.Address) string {
	totalFLIPBalance := flipToken.BalanceOf(user)

	lockedFLIPBalance := getUserLockedBalances(user)
	availableFLIPBalance := totalFLIPBalance - uint64(lockedFLIPBalance.Amount)

	userBalances := &userFLIPBalances{
		lockedBalance:   lockedFLIPBalance.String(),
		availableBalance: strconv.FormatUint(availableFLIPBalance, 10),
	}

	userBalancesJSON, err := userBalances.MarshalJSON()
	if err != nil {
		panic("error in marshalling user FLIP balances")
	}

	return string(userBalancesJSON)
}

func getUserLockedBalances(user std.Address) std.Coin {
	lockedFLIPBalance := std.NewCoin("FLIP", 0)

	v, exists := userFlipNFTs.Get(user.String())
	if exists {
		userLockedBalances, ok := v.([]*lockedVault)
		if !ok {
			return lockedFLIPBalance
		}

		for _, lv := range userLockedBalances {
			lockedFLIPBalance = lockedFLIPBalance.Add(lv.lockedFlip)
		}
	}

	return lockedFLIPBalance
}