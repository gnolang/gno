package grc721

// IGRC2981 follows the Ethereum standard
type IGRC2981 interface {
	// RoyaltyInfo retrieves royalty information for a tokenID and salePrice.
	// It returns the payment address, royalty amount, and an error if any.
	RoyaltyInfo(tokenID TokenID, salePrice int64) (address, int64, error)
}

// RoyaltyInfo represents royalty information for a token.
type RoyaltyInfo struct {
	// PaymentAddress is the address where royalty payment should be sent.
	PaymentAddress address
	// Bps is the royalty rate in basis points (1/10000). Per EIP-2981,
	// royaltyAmount = salePrice * Bps / 10000. Example: Bps = 250 => 2.5%.
	Bps int64
}
