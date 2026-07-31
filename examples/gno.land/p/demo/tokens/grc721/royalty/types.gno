// Package royalty is a stackable GRC721 extension implementing EIP-2981 in basis
// points. Writes are issuer-only so a holder cannot zero out a royalty on a
// token they just received.
package royalty

import (
	"errors"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
)

// FeeDenominator is the EIP-2981 basis-point denominator: bps/10000, 10000 == 100%.
const FeeDenominator = 10000

const RoyaltyUpdateEvent = "RoyaltyUpdate"

var (
	ErrInvalidReceiver   = errors.New("invalid royalty receiver")
	ErrInvalidBps        = errors.New("royalty bps out of range")
	ErrTokenNotMinted    = errors.New("token is not minted")
	ErrMaxBpsRange       = errors.New("maxBps out of range (0..10000)")
	ErrInvalidSalePrice  = errors.New("invalid sale price (negative or would overflow)")
)

var zeroAddress = address("")

// RoyaltyInfo is a royalty policy: send Bps/10000 of a sale to Receiver.
type RoyaltyInfo struct {
	Receiver address
	Bps      int64
}

type storage struct {
	maxBps     int64
	hasDefault bool
	def        RoyaltyInfo
	perToken   avl.Tree // grc721.TokenID -> RoyaltyInfo
}

type Royalty struct {
	core *grc721.Token
	st   *storage
}

type Ledger struct {
	core *grc721.Token
	st   *storage
}

var _ grc721.Extension = (*Ledger)(nil)

var _ grc721.ExtensionView = (*Royalty)(nil)
