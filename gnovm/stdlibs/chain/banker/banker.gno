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

// bankerGetCoins queries the banker module for the coins owned by the given
// addr. It returns the amounts and the denominations in two matching slices
// of the same size (ie., any amounts[i] is specified in the corresponding denoms[i]).
func bankerGetCoins(addr string) (denoms []string, amounts []int64)

// bankerSendCoins uses the banker module to send the coins from the address "from"
// to the address "to". The coins are specified using two slices, denoms and amounts,
// which must be of the same size and specify for each amounts[i], its matching
// denoms[i].
//
// The bt (banker type) parameter will be either of 1 (BankerTypeOriginSend),
// 2 (BankerTypeRealmSend), 3 (BankerTypeRealmIssue).
//
// If the bt is 1, the implementation will only to be allowed to spend coins up
// to the maximum specified in the execution context value OriginSend. Furthermore,
// the already spent amounts should be kept track in OriginSendSpent.
//
// All other validation is already performed by the banker Gno package.
func bankerSendCoins(bt uint8, from, to string, denoms []string, amounts []int64)

// bankerTotalCoin queries the banker module to determine the total supply of
// the coin with the given denom.
func bankerTotalCoin(denom string) int64

// bankerIssueCoin uses the banker module to issue (mint) the given amount of
// the coin with the given denom to the address specified in addr.
func bankerIssueCoin(addr string, denom string, amount int64)

// bankerRemoveCoin uses the banker module to remove (burn) the given amount of
// the coin with the given denom from the address specified in addr.
func bankerRemoveCoin(addr string, denom string, amount int64)

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

// assertCallerIsRealm ensures that the caller of the function calling it has a
// realm pkgPath.
//
// In order to be a realm, the caller's package path must match the following regular expression:
//
//	^(?P<PKGPATH>(?:(?P<DOMAIN>(?:(?P<SLD>(?:(?:(?:[a-z0-9-])+\.)+))(?P<TLD>(?:(?:[a-z]){2,63}))))(?P<URLPATH>(?:/(?P<LETTER>(?:[a-z]))/(?P<USER>(?:(?:(?:_)?[a-z](?:[a-z0-9_])*)))(?:/(?P<REPO>(?:(?:(?:_)?[a-z](?:[a-z0-9_])*)(?:/(?:(?:_)?[a-z](?:[a-z0-9_])*))*)))?))))$
//
// Additionally, the value of the LETTER capture group must be "r", and the REPO
// subgroup must not terminate in "_test".
func assertCallerIsRealm()

// originSend returns the coins specified in the execution context variable
// OriginSend, which specifies the coins sent within the same transaction from
// the caller to the called realm. It returns the amounts and the denominations
// in two matching slices of the same size (ie., any amounts[i] is specified in
// the corresponding denoms[i]).
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
