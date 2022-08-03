package exts

type TokenMetadata interface {
	// Returns the name of the token.
	Name() string

	// Returns the symbol of the token.
	Symbol() string

	// Returns the decimals places of the token.
	Decimals() uint
}
