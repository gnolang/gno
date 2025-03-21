package gnoxchange

import (
	"errors"
	"math"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/r/matijamarjanovic/tokenhub"
)

var (
	ticketCounter uint64
)

func init() {
	ticketCounter = 0
}

// CreateCoinToTokenTicket creates a ticket to swap native coins for GRC20 tokens.
// The coins to be swapped must be sent with the transaction. They will be locked in the realm until
// one of the following happens (this is the case for all tickets):
// - the ticket is fulfilled
// - the ticket is cancelled
// - the ticket expires
func CreateCoinToTokenTicket(
	coinDenom string, // e.g. "ugnot"
	tokenKey string, // e.g. "gno.land/r/test.gtoken"
	minAmountOut uint64, // minimum amount of tokens to receive
	expiryHours int64,
) (string, error) {
	caller := std.PreviousRealm().Address()

	sent := std.OriginSend()
	if len(sent) != 1 || sent[0].Denom != coinDenom {
		return "", errors.New("sent coins don't match ticket parameters")
	}
	amountIn := uint64(sent[0].Amount)

	tokenOut := tokenhub.GetToken(tokenKey)
	if tokenOut == nil {
		return "", errors.New("token not found: " + tokenKey)
	}

	coinDecimals := uint(6)
	tokenDecimals := tokenOut.GetDecimals()

	minDecimals := coinDecimals
	if tokenDecimals < coinDecimals {
		minDecimals = tokenDecimals
	}

	if tokenDecimals > minDecimals {
		expectedScale := uint64(math.Pow10(int(tokenDecimals - minDecimals)))
		if minAmountOut%expectedScale != 0 {
			return "", errors.New(ufmt.Sprintf(
				"invalid minimum output amount: amount must be in units of %d (token has %d decimals vs %d decimals for native coin)",
				expectedScale, tokenDecimals, coinDecimals,
			))
		}
	}

	ticketCounter++
	ticketID := ufmt.Sprintf("swap-%d", ticketCounter)

	ticket := &Ticket{
		ID:           ticketID,
		Creator:      caller,
		AssetIn:      NewCoinAsset(coinDenom),
		AssetOut:     NewTokenAsset(tokenKey),
		AmountIn:     amountIn,
		MinAmountOut: minAmountOut,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		Status:       "open",
	}

	addTicket(ticket)
	return ticketID, nil
}

// CreateTokenToCoinTicket creates a ticket to swap GRC20 tokens for native coins.
// Requires token approval before creating the ticket.
func CreateTokenToCoinTicket(
	tokenKey string, // e.g. "gno.land/r/test.gtoken"
	coinDenom string, // e.g. "ugnot"
	amountIn uint64, // amount of tokens to swap
	minAmountOut uint64, // minimum amount of coins to receive
	expiryHours int64,
) (string, error) {
	caller := std.PreviousRealm().Address()

	tokenIn := tokenhub.GetToken(tokenKey)
	if tokenIn == nil {
		return "", errors.New("token not found: " + tokenKey)
	}

	coinDecimals := uint(6)
	tokenDecimals := tokenIn.GetDecimals()

	minDecimals := coinDecimals
	if tokenDecimals < coinDecimals {
		minDecimals = tokenDecimals
	}

	if tokenDecimals > minDecimals {
		expectedScale := uint64(math.Pow10(int(tokenDecimals - minDecimals)))
		if amountIn%expectedScale != 0 {
			return "", errors.New(ufmt.Sprintf(
				"invalid input amount: amount must be in units of %d (token has %d decimals vs %d decimals for native coin)",
				expectedScale, tokenDecimals, coinDecimals,
			))
		}
	}

	tellerIn := tokenIn.RealmTeller()
	if err := tellerIn.TransferFrom(caller, std.CurrentRealm().Address(), amountIn); err != nil {
		return "", errors.New("failed to transfer tokens: " + err.Error())
	}

	ticketCounter++
	ticketID := ufmt.Sprintf("swap-%d", ticketCounter)

	ticket := &Ticket{
		ID:           ticketID,
		Creator:      caller,
		AssetIn:      NewTokenAsset(tokenKey),
		AssetOut:     NewCoinAsset(coinDenom),
		AmountIn:     amountIn,
		MinAmountOut: minAmountOut,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		Status:       "open",
	}

	addTicket(ticket)
	return ticketID, nil
}

// CreateTokenToTokenTicket creates a ticket to swap one GRC20 token for another.
// Requires token approval before creating the ticket.
func CreateTokenToTokenTicket(
	tokenInKey string, // e.g. "gno.land/r/test1.gtokenA"
	tokenOutKey string, // e.g. "gno.land/r/test2.gtokenB"
	amountIn uint64, // amount of tokens to swap
	minAmountOut uint64, // minimum amount of tokens to receive
	expiryHours int64,
) (string, error) {
	caller := std.PreviousRealm().Address()

	tokenIn := tokenhub.GetToken(tokenInKey)
	if tokenIn == nil {
		return "", errors.New("token not found: " + tokenInKey)
	}

	tokenOut := tokenhub.GetToken(tokenOutKey)
	if tokenOut == nil {
		return "", errors.New("token not found: " + tokenOutKey)
	}

	decimalsIn := tokenIn.GetDecimals()
	decimalsOut := tokenOut.GetDecimals()

	minDecimals := decimalsIn
	if decimalsOut < decimalsIn {
		minDecimals = decimalsOut
	}

	if decimalsIn > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsIn - minDecimals)))
		if amountIn%expectedScale != 0 {
			return "", errors.New(ufmt.Sprintf(
				"invalid input amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				tokenInKey, expectedScale, decimalsIn, minDecimals,
			))
		}
	}

	if decimalsOut > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsOut - minDecimals)))
		if minAmountOut%expectedScale != 0 {
			return "", errors.New(ufmt.Sprintf(
				"invalid minimum output amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				tokenOutKey, expectedScale, decimalsOut, minDecimals,
			))
		}
	}

	tellerIn := tokenIn.RealmTeller()
	if err := tellerIn.TransferFrom(caller, std.CurrentRealm().Address(), amountIn); err != nil {
		return "", errors.New("failed to transfer tokens: " + err.Error())
	}

	ticketCounter++
	ticketID := ufmt.Sprintf("swap-%d", ticketCounter)

	ticket := &Ticket{
		ID:           ticketID,
		Creator:      caller,
		AssetIn:      NewTokenAsset(tokenInKey),
		AssetOut:     NewTokenAsset(tokenOutKey),
		AmountIn:     amountIn,
		MinAmountOut: minAmountOut,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		Status:       "open",
	}

	addTicket(ticket)
	return ticketID, nil
}

// FulfillTicket allows a user to fulfill an open ticket.
// If all checks pass, the ticket is fulfilled and the assets are transferred to both parties.
//
// note: If the ticket is asking for coins, the amountOut must be the exact amount of coins sent.
func FulfillTicket(ticketID string, amountOut uint64) error {
	caller := std.PreviousRealm().Address()

	ticketInterface, exists := ticketRegistry.allTickets.Get(ticketID)
	if !exists {
		return errors.New("ticket not found")
	}

	ticket := ticketInterface.(*Ticket)

	if ticket.Status != "open" {
		return errors.New("ticket is not open")
	}

	if time.Now().After(ticket.ExpiresAt) {
		updateTicketStatus(ticket, "expired")
		return errors.New("ticket has expired")
	}

	if amountOut < ticket.MinAmountOut {
		return errors.New("insufficient output amount")
	}

	if ticket.AssetOut.Type == AssetTypeCoin {
		sent := std.OriginSend()
		if len(sent) != 1 || sent[0].Denom != ticket.AssetOut.Denom || uint64(sent[0].Amount) != amountOut {
			return errors.New("sent coins don't match fulfillment parameters")
		}

		banker := std.NewBanker(std.BankerTypeRealmSend)
		banker.SendCoins(std.CurrentRealm().Address(), ticket.Creator, sent)

		if ticket.AssetIn.Type == AssetTypeCoin {
			banker.SendCoins(
				std.CurrentRealm().Address(),
				caller,
				std.Coins{{ticket.AssetIn.Denom, int64(ticket.AmountIn)}},
			)
		} else {
			tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
			tellerIn := tokenIn.RealmTeller()
			if err := tellerIn.Transfer(caller, ticket.AmountIn); err != nil {
				return errors.New("failed to transfer input token: " + err.Error())
			}
		}
	} else {
		tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
		if tokenOut == nil {
			return errors.New("token not found: " + ticket.AssetOut.Token)
		}

		tellerOut := tokenOut.RealmTeller()
		if err := tellerOut.TransferFrom(caller, ticket.Creator, amountOut); err != nil {
			return errors.New("failed to transfer output token: " + err.Error())
		}

		if ticket.AssetIn.Type == AssetTypeCoin {
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(
				std.CurrentRealm().Address(),
				caller,
				std.Coins{{ticket.AssetIn.Denom, int64(ticket.AmountIn)}},
			)
		} else {
			tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
			tellerIn := tokenIn.RealmTeller()
			if err := tellerIn.Transfer(caller, ticket.AmountIn); err != nil {
				return errors.New("failed to transfer input token: " + err.Error())
			}
		}
	}

	updateTicketStatus(ticket, "fulfilled")
	return nil
}

// CancelTicket allows the creator to cancel their ticket and withdraw the tokens/coins before the ticket is expired.
func CancelTicket(ticketID string) error {
	caller := std.PreviousRealm().Address()

	ticketInterface, exists := ticketRegistry.allTickets.Get(ticketID)
	if !exists {
		return errors.New("ticket not found")
	}

	ticket := ticketInterface.(*Ticket)

	if ticket.Creator != caller {
		return errors.New("only ticket creator can cancel")
	}

	if ticket.Status != "open" {
		return errors.New("ticket is not open")
	}

	if ticket.AssetIn.Type == AssetTypeCoin {
		banker := std.NewBanker(std.BankerTypeRealmSend)
		banker.SendCoins(
			std.CurrentRealm().Address(),
			ticket.Creator,
			std.Coins{{ticket.AssetIn.Denom, int64(ticket.AmountIn)}},
		)
	} else if ticket.AssetIn.Type == AssetTypeNFT {
		nft := tokenhub.GetNFT(ticket.AssetIn.NFTPath)
		if nft == nil {
			return errors.New("NFT not found: " + ticket.AssetIn.NFTPath)
		}

		parts := strings.Split(ticket.AssetIn.NFTPath, ".")
		tokenID := parts[len(parts)-1]

		if err := nft.TransferFrom(std.CurrentRealm().Address(), ticket.Creator, grc721.TokenID(tokenID)); err != nil {
			return errors.New(ufmt.Sprintf("failed to return NFT: %v", err))
		}
	} else {
		tokenIn := tokenhub.GetToken(ticket.AssetIn.Token)
		tellerIn := tokenIn.RealmTeller()
		if err := tellerIn.Transfer(ticket.Creator, ticket.AmountIn); err != nil {
			return errors.New(ufmt.Sprintf("failed to refund tokens: %v", err))
		}
	}

	updateTicketStatus(ticket, "cancelled")
	return nil
}

// Helper function to add a ticket to the appropriate trees
func addTicket(ticket *Ticket) {
	ticketID := ticket.ID

	ticketRegistry.allTickets.Set(ticketID, ticket)

	if ticket.Status == "open" && !time.Now().After(ticket.ExpiresAt) {
		if ticket.AssetIn.Type == AssetTypeNFT {
			ticketRegistry.openNFTTickets.Set(ticketID, ticket)
		} else {
			ticketRegistry.openTickets.Set(ticketID, ticket)
		}
	}
}

// Helper function to update ticket status
func updateTicketStatus(ticket *Ticket, newStatus string) {
	oldStatus := ticket.Status
	ticket.Status = newStatus

	ticketRegistry.allTickets.Set(ticket.ID, ticket)

	if oldStatus == "open" && newStatus != "open" {
		if ticket.AssetIn.Type == AssetTypeNFT {
			ticketRegistry.openNFTTickets.Remove(ticket.ID)
		} else {
			ticketRegistry.openTickets.Remove(ticket.ID)
		}
	}
}
