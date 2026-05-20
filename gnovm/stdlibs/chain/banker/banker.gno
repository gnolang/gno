package banker

import (
	"chain"
	"strconv"
	"strings"
)

// Realm functions can call banker.NewBanker(options) to get a banker
// instance. A banker instance can be passed onto other realm functions
// (this allows other realms to spend coins on behalf of the first
// realm) and can be stored in package-level vars across transactions —
// the underlying struct (BankerType + scoped pkgAddr) persists like any
// other Gno value, and the methods continue to dispatch through the
// runtime against live bank state on each call. See
// gno.land/pkg/integration/testdata/banker_persistence.txtar for the
// full cross-realm capability-persistence behavior.
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

// NewReadonlyBanker returns a banker that can only read coin state.
// Readonly bankers have no authority to mutate anything (no SendCoins,
// IssueCoin, or RemoveCoin) so they do not require a realm capability
// — anyone, including Render functions and non-crossing helpers, can
// construct one. Use this in queries; use NewBanker for any operation
// that moves coins.
func NewReadonlyBanker() Banker {
	return banker{bt: BankerTypeReadonly}
}

// NewBanker returns a new Banker bound to rlm — anyone holding rlm can
// construct a banker for it. The realm value is the authorization
// capability: a function that takes `cur realm` (crossing function, or
// a non-first-arg-threaded function) can call NewBanker; a function
// without a realm handle in scope cannot.
//
// bt must be one of BankerTypeOriginSend, BankerTypeRealmSend, or
// BankerTypeRealmIssue. Use NewReadonlyBanker for BankerTypeReadonly —
// the readonly path requires no realm capability.
func NewBanker(bt BankerType, rlm realm) Banker {
	if bt == BankerTypeReadonly {
		panic("use NewReadonlyBanker for BankerTypeReadonly")
	}
	if bt >= maxBanker {
		panic("invalid banker type")
	}
	if bt == BankerTypeOriginSend && !rlm.Previous().IsUserCall() {
		panic("banker with type BankerTypeOriginSend can only be instantiated by the origin package")
	}
	return banker{
		bt:      bt,
		pkgAddr: rlm.Address(),
		pkgPath: rlm.PkgPath(),
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
	pkgPath string // realm pkgpath, for IssueCoin/RemoveCoin denom-prefix.
}

// IsCanonical reports whether b is the canonical bank-keeper-backed
// Banker produced by NewBanker. The concrete `banker` type is unexported,
// so only this package can construct values that satisfy this assertion;
// hand-rolled Banker implementations (no-op fakes, mocks, decorators)
// return false. Use this in cap-token wrappers (e.g. AllowanceSender) to
// reject untrusted Banker substitutions at construction time, before
// any state mutation depends on the banker actually moving coins.
//
// Forward-compat: if this package ever introduces additional canonical
// Banker implementations alongside `banker`, this function MUST be
// updated to recognize them. New bank-keeper-backed types must be added
// to the type-switch explicitly — the current single-type assertion is
// a deliberate allowlist, not an inadvertent restriction.
func IsCanonical(b Banker) bool {
	_, ok := b.(banker)
	return ok
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
	assertCoinDenom(denom, b.pkgPath)
	bankerIssueCoin(uint8(b.bt), string(addr), denom, amount)
}

func (b banker) RemoveCoin(addr address, denom string, amount int64) {
	if b.bt != BankerTypeRealmIssue {
		panic(b.bt.String() + " cannot remove coins")
	}
	assertCoinDenom(denom, b.pkgPath)
	bankerRemoveCoin(uint8(b.bt), string(addr), denom, amount)
}

// assertCoinDenom checks that denom carries the prefix bound to the realm
// the banker was constructed for.
func assertCoinDenom(denom string, pkgPath string) {
	prefix := "/" + pkgPath + ":"
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
