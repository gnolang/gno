package bank

const (
	ModuleName = "bank"

	// SupplyStoreKeyPrefix is the prefix for per-denomination total supply entries.
	SupplyStoreKeyPrefix = "/s/"
)

// SupplyStoreKey returns the store key for a denomination's total supply.
func SupplyStoreKey(denom string) []byte {
	return append([]byte(SupplyStoreKeyPrefix), []byte(denom)...)
}
