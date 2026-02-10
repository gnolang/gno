package grc721

// IGRC2981 follows the Ethereum standard
type IGRC2981 interface {
	// RoyaltyInfo retrieves royalty information for a tokenID and salePrice.
	// It returns the payment address, royalty amount, and an error if any.
	RoyaltyInfo(tokenID TokenID, salePrice int64) (address, int64, error)
}

// RoyaltyInfo represents royalty information for a token.
type RoyaltyInfo struct {
	PaymentAddress address // PaymentAddress is the address where royalty payment should be sent.
	Percentage     int64   // Percentage is the royalty percentage. It indicates the percentage of royalty to be paid for each sale. For example : Percentage = 10 => 10%
}
