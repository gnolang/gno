package flippando

import "errors"

var (
	ErrUserHasZeroGame             	= errors.New("user has zero games")
	ErrSVGDataFailed				= errors.New("SVG data not generated")
	ErrGameNotComplete				= errors.New("game not complete, board has gaps")
	ErrInvalidTokenId				= errors.New("invalid TokenID")
	ErrUnlockingRecipientNotFound	= errors.New("recipient for unlocking not found")
	ErrMarketplaceListingDoesNotExist	= errors.New("This listing does not exist")
	ErrFlipTokenBalanceOfFailed			= errors.New("Error in calling BalanceOf for FLIP")
	ErrFlipTokenNotEnoughBalance		= errors.New("Not enough FLIP balance for this trade")
)