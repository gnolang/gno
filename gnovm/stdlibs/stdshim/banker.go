package std

// Realm functions can call std.GetBanker(options) to get
// a banker instance. Banker objects cannot be persisted,
// but can be passed onto other functions to be transacted
// on. A banker instance can be passed onto other realm
// functions; this allows other realms to spend coins on
// behalf of the first realm.
//
// Banker panics on errors instead of returning errors.
// This also helps simplify the interface and prevent
// hidden bugs (e.g. ignoring errors)
//
// NOTE: this Gno interface is satisfied by a native go
// type, and those can't return non-primitive objects
// (without confusion).
type Banker interface {
	GetCoins(addr Address) (dst Coins)
	SendCoins(from, to Address, amt Coins)
	TotalCoin(denom string) int64
	IssueCoin(addr Address, denom string, amount int64)
	RemoveCoin(addr Address, denom string, amount int64)
}

// Also available natively in stdlibs/context.go
type BankerType uint8

// Also available natively in stdlibs/context.go
const (
	// Can only read state.
	BankerTypeReadonly BankerType = iota
	// Can only send from tx send.
	BankerTypeOrigSend
	// Can send from all realm coins.
	BankerTypeRealmSend
	// Can issue and remove realm coins.
	BankerTypeRealmIssue
)

//----------------------------------------
// adapter for native banker

type bankAdapter struct {
	nativeBanker Banker
}

func (ba bankAdapter) GetCoins(addr Address) (dst Coins) {
	// convert native -> gno
	coins := ba.nativeBanker.GetCoins(addr)
	for _, coin := range coins {
		dst = append(dst, (Coin)(coin))
	}
	return dst
}

func (ba bankAdapter) SendCoins(from, to Address, amt Coins) {
	ba.nativeBanker.SendCoins(from, to, amt)
}

func (ba bankAdapter) TotalCoin(denom string) int64 {
	return ba.nativeBanker.TotalCoin(denom)
}

func (ba bankAdapter) IssueCoin(addr Address, denom string, amount int64) {
	ba.nativeBanker.IssueCoin(addr, denom, amount)
}

func (ba bankAdapter) RemoveCoin(addr Address, denom string, amount int64) {
	ba.nativeBanker.RemoveCoin(addr, denom, amount)
}
