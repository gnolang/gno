package banker

import (
	"chain"
	"chain/runtime"
	"strconv"
	"strings"
)

// Realm functions can call std.NewBanker(options) to get
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
	GetCoins(addr address) (dst chain.Coins)
	SendCoins(from, to address, amt chain.Coins)
	TotalCoin(denom string) int64
	IssueCoin(addr address, denom string, amount int64)
	RemoveCoin(addr address, denom string, amount int64)
}

// BankerType represents the "permission level" requested for a banker,
// retrievable through [NewBanker].
type BankerType uint8

// Available types of banker.
const (
	// Can only read state.
	BankerTypeReadonly BankerType = iota
	// Can only send from tx send.
	BankerTypeOriginSend
	// Can send from all realm coins.
	BankerTypeRealmSend
	// Can issue and remove realm coins.
	BankerTypeRealmIssue

	maxBanker
)

func (b BankerType) String() string {
	switch b {
	case BankerTypeReadonly:
		return "BankerTypeReadonly"
	case BankerTypeOriginSend:
		return "BankerTypeOriginSend"
	case BankerTypeRealmSend:
		return "BankerTypeRealmSend"
	case BankerTypeRealmIssue:
		return "BankerTypeRealmIssue"
	default:
		panic("invalid BankerType: " + strconv.Itoa(int(b)))
	}
}

//----------------------------------------
// adapter for native banker

// NewBanker returns a new Banker, with its capabilities matching the given
// [BankerType].
func NewBanker(bt BankerType) Banker {
	assertCallerIsRealm()
	if bt >= maxBanker {
		panic("invalid banker type")
	}

	var pkgAddr address
	if bt == BankerTypeOriginSend {
		pkgAddr = runtime.CurrentRealm().Address()
		if runtime.PreviousRealm().PkgPath() != "" {
			panic("banker with type BankerTypeOriginSend can only be instantiated by the origin package")
		}
	} else if bt == BankerTypeRealmSend || bt == BankerTypeRealmIssue {
		pkgAddr = runtime.CurrentRealm().Address()
	}
	return banker{
		bt,
		pkgAddr,
	}
}

// These are native bindings to the banker's functions.
func bankerGetCoins(bt uint8, addr string) (denoms []string, amounts []int64)
func bankerSendCoins(bt uint8, from, to string, denoms []string, amounts []int64)
func bankerTotalCoin(bt uint8, denom string) int64
func bankerIssueCoin(bt uint8, addr string, denom string, amount int64)
func bankerRemoveCoin(bt uint8, addr string, denom string, amount int64)

type banker struct {
	bt      BankerType
	pkgAddr address
}

func (b banker) GetCoins(addr address) (dst chain.Coins) {
	denoms, amounts := bankerGetCoins(uint8(b.bt), string(addr))
	dst = make(chain.Coins, len(denoms))
	for i := range dst {
		dst[i] = chain.Coin{denoms[i], amounts[i]}
	}
	return dst
}

func (b banker) SendCoins(from, to address, amt chain.Coins) {
	if b.bt == BankerTypeReadonly {
		panic("BankerTypeReadonly cannot send coins")
	}
	if b.pkgAddr != from {
		msg := `can only send coins from realm that created banker "` + b.pkgAddr + `", not "` + from + `"`
		panic(msg)
	}
	denoms, amounts := expandNative(amt)
	bankerSendCoins(uint8(b.bt), string(from), string(to), denoms, amounts)
}

func (b banker) TotalCoin(denom string) int64 {
	return bankerTotalCoin(uint8(b.bt), denom)
}

func (b banker) IssueCoin(addr address, denom string, amount int64) {
	if b.bt != BankerTypeRealmIssue {
		panic(b.bt.String() + " cannot issue coins")
	}
	assertCoinDenom(denom)
	bankerIssueCoin(uint8(b.bt), string(addr), denom, amount)
}

func (b banker) RemoveCoin(addr address, denom string, amount int64) {
	if b.bt != BankerTypeRealmIssue {
		panic(b.bt.String() + " cannot remove coins")
	}
	assertCoinDenom(denom)
	bankerRemoveCoin(uint8(b.bt), string(addr), denom, amount)
}

func assertCoinDenom(denom string) {
	prefix := "/" + runtime.CurrentRealm().PkgPath() + ":"
	if !strings.HasPrefix(denom, prefix) {
		panic("invalid denom, can only issue/remove coins with the realm's prefix: " + prefix)
	}

	baseDenom := denom[len(prefix):]
	if !isValidBaseDenom(baseDenom) {
		panic("cannot issue coins with invalid denom base name, it should start by a lowercase letter and be followed by 2-15 lowercase letters or digits")
	}
}

// check start by a lowercase letter and be followed by 2-15 lowercase letters or digits
func isValidBaseDenom(denom string) bool {
	length := len(denom)
	if length < 3 || length > 16 {
		return false
	}
	for i, c := range denom {
		switch {
		case c >= 'a' && c <= 'z',
			i > 0 && (c >= '0' && c <= '9'): // continue
		default:
			return false
		}
	}
	return true
}

func OriginSend() chain.Coins {
	den, amt := originSend()
	coins := make(chain.Coins, len(den))
	for i := range coins {
		coins[i] = chain.Coin{Denom: den[i], Amount: amt[i]}
	}
	return coins
}

func assertCallerIsRealm()
func originSend() (denoms []string, amounts []int64)

// expandNative expands coins for usage within natively bound functions.
func expandNative(cz chain.Coins) (denoms []string, amounts []int64) {
	denoms = make([]string, len(cz))
	amounts = make([]int64, len(cz))
	for i, coin := range cz {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}

	return denoms, amounts
}
