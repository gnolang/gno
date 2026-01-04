package fomo3d

import "errors"

var (
	// Game state errors
	ErrGameInProgress    = errors.New("fomo3d: game already in progress")
	ErrGameNotInProgress = errors.New("fomo3d: game not in progress")
	ErrGameEnded         = errors.New("fomo3d: game has ended")
	ErrGameTimeExpired   = errors.New("fomo3d: game time expired")
	ErrNoKeysPurchased   = errors.New("fomo3d: no keys purchased")
	ErrPlayerNotInGame   = errors.New("fomo3d: player is not in the game")

	// Payment errors
	ErrInvalidPayment      = errors.New("fomo3d: must send ugnot only")
	ErrInsufficientPayment = errors.New("fomo3d: insufficient payment for key")

	// Dividend errors
	ErrNoDividendsToClaim = errors.New("fomo3d: no dividends to claim")

	// Fee errors
	ErrNoFeesToClaim = errors.New("fomo3d: no owner fees to claim")

	// Resolution errors
	ErrInvalidAddressOrName = errors.New("fomo3d: invalid address or unregistered username")

	// NFT errors
	ErrUnauthorizedMint = errors.New("fomo3d: only the Fomo3D game realm can mint winner NFTs")
	ErrZeroAddress      = errors.New("fomo3d: zero address")
)
