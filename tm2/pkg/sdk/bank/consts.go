package bank

const (
	ModuleName = "bank"

	// TotalCoinStoreKeyPrefix prefix for total-coin-by-denom store
	TotalCoinStoreKeyPrefix = "/tc/"
)

// TotalCoinStoreKey turn an denom to key used to get it from the total coin store
func TotalCoinStoreKey(denom string) []byte {
	return []byte(TotalCoinStoreKeyPrefix + denom)
}
