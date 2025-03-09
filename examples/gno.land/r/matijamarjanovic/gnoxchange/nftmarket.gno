package gnoxchange

import (
	"errors"
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
		return "", errors.New("invalid NFT path format: expected path.collection.tokenID")
	}

	tokenID := parts[len(parts)-1]

	nft := tokenhub.GetNFT(nftFullPath)
	if nft == nil {
		return "", errors.New("NFT not found: " + nftFullPath)
	}

	owner, err := nft.OwnerOf(grc721.TokenID(tokenID))
	if err != nil {
		return "", errors.New("invalid token ID")
	}
	if owner != caller {
		return "", errors.New("caller is not the owner of the NFT")
	}

	tokenOut := tokenhub.GetToken(tokenKey)
	if tokenOut == nil {
		return "", errors.New("token not found: " + tokenKey)
	}

	if err := nft.TransferFrom(caller, std.CurrentRealm().Address(), grc721.TokenID(tokenID)); err != nil {
		return "", errors.New("failed to transfer NFT: " + err.Error())
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
		return "", errors.New("invalid NFT path format: expected path.collection.tokenID")
	}

	tokenID := parts[len(parts)-1]

	nft := tokenhub.GetNFT(nftFullPath)
	if nft == nil {
		return "", errors.New("NFT not found: " + nftFullPath)
	}

	owner, err := nft.OwnerOf(grc721.TokenID(tokenID))
	if err != nil {
		return "", errors.New("invalid token ID")
	}
	if owner != caller {
		return "", errors.New("caller is not the owner of the NFT")
	}

	if err := nft.TransferFrom(caller, std.CurrentRealm().Address(), grc721.TokenID(tokenID)); err != nil {
		return "", errors.New("failed to transfer NFT: " + err.Error())
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
		return errors.New("ticket not found")
	}

	ticket := ticketInterface.(*Ticket)

	if time.Now().After(ticket.ExpiresAt) {
		updateTicketStatus(ticket, "expired")
		return errors.New("ticket has expired")
	}

	if amountOut < ticket.MinAmountOut {
		return errors.New("insufficient payment amount")
	}

	if ticket.AssetIn.Type != AssetTypeNFT {
		return errors.New("not an NFT ticket")
	}

	nft := tokenhub.GetNFT(ticket.AssetIn.NFTPath)
	if nft == nil {
		return errors.New("NFT not found: " + ticket.AssetIn.NFTPath)
	}

	if ticket.AssetOut.Type == AssetTypeCoin {
		sent := std.OriginSend()
		if len(sent) != 1 || sent[0].Denom != ticket.AssetOut.Denom || uint64(sent[0].Amount) != amountOut {
			return errors.New("sent coins don't match payment parameters")
		}

		banker := std.NewBanker(std.BankerTypeRealmSend)
		banker.SendCoins(std.CurrentRealm().Address(), ticket.Creator, sent)
	} else {
		tokenOut := tokenhub.GetToken(ticket.AssetOut.Token)
		if tokenOut == nil {
			return errors.New("token not found: " + ticket.AssetOut.Token)
		}

		tellerOut := tokenOut.RealmTeller()
		if err := tellerOut.TransferFrom(caller, ticket.Creator, amountOut); err != nil {
			return errors.New("failed to transfer payment tokens: " + err.Error())
		}
	}
	nftKeyParts := strings.Split(ticket.AssetIn.NFTPath, ".")
	tokenID := nftKeyParts[len(nftKeyParts)-1]

	if err := nft.TransferFrom(std.CurrentRealm().Address(), caller, grc721.TokenID(tokenID)); err != nil {
		panic(ufmt.Sprintf("CRITICAL: payment processed but NFT transfer failed: %v", err))
	}

	updateTicketStatus(ticket, "fulfilled")
	return nil
}
