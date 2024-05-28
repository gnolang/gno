package std

import (
	"fmt"
	"regexp"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// BankerInterface is the interface through which Gno is capable of accessing
// the blockchain's banker.
//
// The name is what it is to avoid a collision with Gno's Banker, when
// transpiling.
type BankerInterface interface {
	GetCoins(addr crypto.Bech32Address) (dst std.Coins)
	SendCoins(from, to crypto.Bech32Address, amt std.Coins)
	TotalCoin(denom string) int64
	IssueCoin(addr crypto.Bech32Address, denom string, amount int64)
	RemoveCoin(addr crypto.Bech32Address, denom string, amount int64)
}

const (
	// Can only read state.
	btReadonly uint8 = iota //nolint
	// Can only send from tx send.
	btOrigSend
	// Can send from all realm coins.
	btRealmSend
	// Can issue and remove realm coins.
	btRealmIssue
)

// regexp for denom format
var reDenom = regexp.MustCompile("[a-z][a-z0-9]{2,15}")

func X_bankerGetCoins(m *gno.Machine, bt uint8, addr string) (denoms []string, amounts []int64) {
	coins := m.Context.(ExecContext).Banker.GetCoins(crypto.Bech32Address(addr))
	return ExpandCoins(coins)
}

func X_bankerSendCoins(m *gno.Machine, bt uint8, fromS, toS string, denoms []string, amounts []int64) {
	// bt != BankerTypeReadonly (checked in gno)

	ctx := m.Context.(ExecContext)
	amt := CompactCoins(denoms, amounts)
	from, to := crypto.Bech32Address(fromS), crypto.Bech32Address(toS)

	pkgAddr := ctx.OrigPkgAddr
	if m.Realm != nil {
		pkgPath := m.Realm.Path
		pkgAddr = gno.DerivePkgAddr(pkgPath).Bech32()
	}

	if bt == btOrigSend || bt == btRealmSend {
		if from != pkgAddr {
			m.Panic(typedString(
				fmt.Sprintf(
					"can only send from the realm package address %q, but got %q",
					pkgAddr, from),
			))
			return
		}
	}

	switch bt {
	case btOrigSend:
		// indirection allows us to "commit" in a second phase
		spent := (*ctx.OrigSendSpent).Add(amt)
		if !ctx.OrigSend.IsAllGTE(spent) {
			m.Panic(typedString(
				fmt.Sprintf(
					`cannot send "%v", limit "%v" exceeded with "%v" already spent`,
					amt, ctx.OrigSend, *ctx.OrigSendSpent),
			))
			return
		}
		ctx.Banker.SendCoins(from, to, amt)
		*ctx.OrigSendSpent = spent
	case btRealmSend, btRealmIssue:
		ctx.Banker.SendCoins(from, to, amt)
	default:
		panic(fmt.Sprintf("invalid banker type %d in bankerSendCoins", bt))
	}
}

func X_bankerTotalCoin(m *gno.Machine, bt uint8, denom string) int64 {
	return m.Context.(ExecContext).Banker.TotalCoin(denom)
}

func X_bankerIssueCoin(m *gno.Machine, bt uint8, addr string, denom string, amount int64) {
	// gno checks for bt == RealmIssue

	// check origin denom format
	matched := reDenom.MatchString(denom)
	if !matched {
		m.Panic(typedString("invalid denom format to issue coin, must be " + reDenom.String()))
		return
	}

	// Similar to ibc spec
	// ibc_denom := 'ibc/' + hash('path' + 'base_denom')
	// gno_realm_denom := '/' + 'pkg_path' + ':' + 'base_denom'
	newDenom := "/" + m.Realm.Path + ":" + denom
	m.Context.(ExecContext).Banker.IssueCoin(crypto.Bech32Address(addr), newDenom, amount)
}

func X_bankerRemoveCoin(m *gno.Machine, bt uint8, addr string, denom string, amount int64) {
	// gno checks for bt == RealmIssue

	matched := reDenom.MatchString(denom)
	if !matched {
		m.Panic(typedString("invalid denom format to remove coin, must be " + reDenom.String()))
		return
	}

	newDenom := "/" + m.Realm.Path + ":" + denom
	m.Context.(ExecContext).Banker.RemoveCoin(crypto.Bech32Address(addr), newDenom, amount)
}
