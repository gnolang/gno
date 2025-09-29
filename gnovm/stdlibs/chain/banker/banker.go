package banker

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
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
	btOriginSend
	// Can send from all realm coins.
	btRealmSend
	// Can issue and remove realm coins.
	btRealmIssue
)

func X_bankerGetCoins(m *gno.Machine, bt uint8, addr string) (denoms []string, amounts []int64) {
	coins := execctx.GetContext(m).Banker.GetCoins(crypto.Bech32Address(addr))
	return ExpandCoins(coins)
}

func X_bankerSendCoins(m *gno.Machine, bt uint8, fromS, toS string, denoms []string, amounts []int64) {
	// bt != BankerTypeReadonly (checked in gno)

	ctx := execctx.GetContext(m)
	amt := CompactCoins(denoms, amounts)
	from, to := crypto.Bech32Address(fromS), crypto.Bech32Address(toS)

	switch bt {
	case btOriginSend:
		// indirection allows us to "commit" in a second phase
		spent := (*ctx.OriginSendSpent).Add(amt)
		if !ctx.OriginSend.IsAllGTE(spent) {
			m.PanicString(
				fmt.Sprintf(
					`cannot send "%v", limit "%v" exceeded with "%v" already spent`,
					amt, ctx.OriginSend, *ctx.OriginSendSpent),
			)
			return
		}
		ctx.Banker.SendCoins(from, to, amt)
		*ctx.OriginSendSpent = spent
	case btRealmSend, btRealmIssue:
		ctx.Banker.SendCoins(from, to, amt)
	default:
		panic(fmt.Sprintf("invalid banker type %d in bankerSendCoins", bt))
	}
}

func X_bankerTotalCoin(m *gno.Machine, bt uint8, denom string) int64 {
	return execctx.GetContext(m).Banker.TotalCoin(denom)
}

func X_bankerIssueCoin(m *gno.Machine, bt uint8, addr string, denom string, amount int64) {
	execctx.GetContext(m).Banker.IssueCoin(crypto.Bech32Address(addr), denom, amount)
}

func X_bankerRemoveCoin(m *gno.Machine, bt uint8, addr string, denom string, amount int64) {
	execctx.GetContext(m).Banker.RemoveCoin(crypto.Bech32Address(addr), denom, amount)
}

func ExpandCoins(c std.Coins) (denoms []string, amounts []int64) {
	denoms = make([]string, len(c))
	amounts = make([]int64, len(c))
	for i, coin := range c {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}
	return denoms, amounts
}

func CompactCoins(denoms []string, amounts []int64) std.Coins {
	coins := make(std.Coins, len(denoms))
	for i := range coins {
		coins[i] = std.Coin{Denom: denoms[i], Amount: amounts[i]}
	}
	return coins
}

func X_assertCallerIsRealm(m *gno.Machine) {
	frame := m.Frames[m.NumFrames()-2]
	if path := frame.LastPackage.PkgPath; !gno.IsRealmPath(path) {
		m.PanicString("caller is not a realm")
	}
}

func X_originSend(m *gno.Machine) (denoms []string, amounts []int64) {
	os := execctx.GetContext(m).OriginSend
	return ExpandCoins(os)
}
