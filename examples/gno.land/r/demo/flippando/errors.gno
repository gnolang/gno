package flippando

import "errors"

var (
	ErrUserHasZeroGame             	= errors.New("user has zero games")
	ErrSVGDataFailed				= errors.New("SVG data not generated")
	ErrGameNotComplete				= errors.New("game not complete, board has gaps")
	ErrInvalidTokenId				= errors.New("invalid TokenID")
	ErrUnlockingRecipientNotFound	= errors.New("recipient for unlocking not found")
)