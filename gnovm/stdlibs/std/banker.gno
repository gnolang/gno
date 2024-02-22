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

// BankerType represents the "permission level" requested for a banker,
// retrievable through [GetBanker].
type BankerType uint8

// Available types of banker.
const (
	// Can only read state.
	BankerTypeReadonly BankerType = iota
	// Can only send from tx send.
	BankerTypeOrigSend
	// Can send from all realm coins.
	BankerTypeRealmSend
	// Can issue and remove realm coins.
	BankerTypeRealmIssue

	maxBanker
)

//----------------------------------------
// adapter for native banker

// GetBanker returns a new Banker, with its capabilities matching the given
// [BankerType].
func GetBanker(bt BankerType) Banker {
	if bt >= maxBanker {
		panic("invalid banker type")
	}
	return banker{bt}
}

// These are native bindings to the banker's functions.
func bankerGetCoins(bt uint8, addr string) (denoms []string, amounts []int64)
func bankerSendCoins(bt uint8, from, to string, denoms []string, amounts []int64)
func bankerTotalCoin(bt uint8, denom string) int64
func bankerIssueCoin(bt uint8, addr string, denom string, amount string)
func bankerRemoveCoin(bt uint8, addr string, denom string, amount string)

type banker struct {
	bt uint8
}

func (b banker) GetCoins(addr Address) (dst Coins) {
	denoms, amounts := bankerGetCoins(b.bt, string(addr))
	dst = make(Coins, len(denoms))
	for i := range dst {
		dst[i] = Coin{denoms[i], amounts[i]}
	}
	return dst
}

func (b banker) SendCoins(from, to Address, amt Coins) {
	if b.bt == BankerTypeReadonly {
		panic("BankerTypeReadonly cannot send coins")
	}
	denoms := make([]string, len(amt))
	amounts := make([]int64, len(amt))
	for i, coin := range amt {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}
	bankerSendCoins(b.bt, string(from), string(to), denoms, amounts)
}

func (b banker) TotalCoin(denom string) int64 {
	return bankerTotalCoin(b.bt, denom)
}

func (b banker) IssueCoin(addr Address, denom string, amount int64) {
	if b.bt == BankerTypeReadonly {
		panic("BankerTypeReadonly cannot issue coins")
	}
	bankerIssueCoin(b.bt, string(addr), denom, amount)
}

func (b banker) RemoveCoin(addr Address, denom string, amount int64) {
	if b.bt == BankerTypeReadonly {
		panic("BankerTypeReadonly cannot remove coins")
	}
	bankerRemoveCoin(b.bt, string(addr), denom, amount)
}
