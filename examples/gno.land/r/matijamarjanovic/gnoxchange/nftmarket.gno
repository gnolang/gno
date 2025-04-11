package gnoxchange

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/r/matijamarjanovic/tokenhub"
)

// CreateNFTToTokenTicket creates a ticket to swap a GRC721 NFT for GRC20 tokens.
// Requires NFT approval before creating the ticket.
func CreateNFTToTokenTicket(
	nftFullPath string, // e.g. "gno.land/r/test.nft.mycollection.1"
	tokenKey string, // e.g. "gno.land/r/test.gtoken"
	minAmountOut uint64, // minimum amount of tokens to receive
	expiryHours int64,
) (string, error) {
	caller := std.PreviousRealm().Address()

	parts := strings.Split(nftFullPath, ".")
	if len(parts) < 4 {
		return "", ErrInvalidNFTPathFormat
	}

	tokenID := parts[len(parts)-1]

	nft := tokenhub.GetNFT(nftFullPath)
	if nft == nil {
		return "", ErrNFTNotFound(nftFullPath)
	}

	owner, err := nft.OwnerOf(grc721.TokenID(tokenID))
	if err != nil {
		return "", ErrInvalidTokenID
	}
	if owner != caller {
		return "", ErrCallerNotOwner
	}

	tokenOut := tokenhub.GetToken(tokenKey)
	if tokenOut == nil {
		return "", ErrTokenNotFound(tokenKey)
	}

	tokenDecimals := tokenOut.GetDecimals()
	if tokenDecimals > 6 {
		return "", ErrTokenDecimalsExceedRange(tokenKey, tokenDecimals)
	}

	if err := nft.TransferFrom(caller, std.CurrentRealm().Address(), grc721.TokenID(tokenID)); err != nil {
		return "", ErrFailedNFTTransfer(err)
	}

	ticketCounter++
	ticketID := ufmt.Sprintf("nfts-%d", ticketCounter)

	ticket := &Ticket{
		ID:           ticketID,
		Creator:      caller,
		AssetIn:      NewNFTAsset(nftFullPath),
		AssetOut:     NewTokenAsset(tokenKey),
		AmountIn:     1, // non-fungible
		MinAmountOut: minAmountOut,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		Status:       "open",
	}

	addTicket(ticket)
	return ticketID, nil
}

// CreateNFTToCoinTicket creates a ticket to swap a GRC721 NFT for native coins.
func CreateNFTToCoinTicket(
	nftFullPath string, // e.g. "gno.land/r/test.nft.mycollection.1"
	coinDenom string, // e.g. "ugnot"
	minAmountOut uint64, // minimum amount of coins to receive
	expiryHours int64,
) (string, error) {
	caller := std.PreviousRealm().Address()

	parts := strings.Split(nftFullPath, ".")
	if len(parts) < 4 {
		return "", ErrInvalidNFTPathFormat
	}

	tokenID := parts[len(parts)-1]

	nft := tokenhub.GetNFT(nftFullPath)
	if nft == nil {
		return "", ErrNFTNotFound(nftFullPath)
	}

	owner, err := nft.OwnerOf(grc721.TokenID(tokenID))
	if err != nil {
		return "", ErrInvalidTokenID
	}
	if owner != caller {
		return "", ErrCallerNotOwner
	}

	if err := nft.TransferFrom(caller, std.CurrentRealm().Address(), grc721.TokenID(tokenID)); err != nil {
		return "", ErrFailedNFTTransfer(err)
	}

	ticketCounter++
	ticketID := ufmt.Sprintf("nfts-%d", ticketCounter)

	ticket := &Ticket{
		ID:           ticketID,
		Creator:      caller,
		AssetIn:      NewNFTAsset(nftFullPath),
		AssetOut:     NewCoinAsset(coinDenom),
		AmountIn:     1, // non-fungible
		MinAmountOut: minAmountOut,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		Status:       "open",
	}

	addTicket(ticket)
	return ticketID, nil
}

// BuyNFT allows a user to fulfill an NFT ticket by paying with tokens or coins.
// It functions on the same principle as the fulfillTicket function.
func BuyNFT(ticketID string, amountOut uint64) error {
	caller := std.PreviousRealm().Address()

	ticketInterface, exists := ticketRegistry.openNFTTickets.Get(ticketID)
	if !exists {
		return ErrTicketNotFound
	}

	ticket := ticketInterface.(*Ticket)

	if time.Now().After(ticket.ExpiresAt) {
		updateTicketStatus(ticket, "expired")
		return ErrTicketExpired
	}

	if amountOut < ticket.MinAmountOut {
		return ErrInsufficientOutputAmount(ticket.MinAmountOut, amountOut)
	}

	if ticket.AssetIn.Type != AssetTypeNFT {
		return ErrInvalidInputToken
	}

	nft := tokenhub.GetNFT(ticket.AssetIn.NFTPath)
	if nft == nil {
		return ErrNFTNotFound(ticket.AssetIn.NFTPath)
	}

	if ticket.AssetOut.Type == AssetTypeCoin {
		sent := std.OriginSend()
		if len(sent) != 1 || sent[0].Denom != ticket.AssetOut.Denom || uint64(sent[0].Amount) != amountOut {
			return ErrSentCoinsMismatch
		}

		banker := std.NewBanker(std.BankerTypeRealmSend)
		banker.SendCoins(std.CurrentRealm().Address(), ticket.Creator, sent)
	} else {
		tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
		if tokenOut == nil {
			return ErrTokenNotFound(ticket.AssetOut.Token)
		}

		tellerOut := tokenOut.RealmTeller()
		if err := tellerOut.TransferFrom(caller, ticket.Creator, amountOut); err != nil {
			return ErrFailedTokenTransfer(ticket.AssetOut.Token, err)
		}
	}
	nftKeyParts := strings.Split(ticket.AssetIn.NFTPath, ".")
	tokenID := nftKeyParts[len(nftKeyParts)-1]

	if err := nft.TransferFrom(std.CurrentRealm().Address(), caller, grc721.TokenID(tokenID)); err != nil {
		return ErrFailedTokenTransfer(ticket.AssetIn.NFTPath, err)
	}

	updateTicketStatus(ticket, "fulfilled")
	return nil
}
