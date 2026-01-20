package bank

const (
	ModuleName     = "bank"
	StoreKeyPrefix = "/bk/"
)

func storeKey(key string) []byte {
	return append([]byte(StoreKeyPrefix), []byte(key)...)
}
