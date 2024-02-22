package std

import (
	"fmt"

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
	btReadonly uint8 = iota
	// Can only send from tx send.
	btOrigSend
	// Can send from all realm coins.
	btRealmSend
	// Can issue and remove realm coins.
	btRealmIssue
)

func X_bankerGetCoins(m *gno.Machine, bt uint8, addr string) (denoms []string, amounts []int64) {
	coins := m.Context.(ExecContext).Banker.GetCoins(crypto.Bech32Address())
	denoms = make([]string, len(coins))
	amounts = make([]string, len(coins))
	for i, coin := range coins {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amounts
	}
	return denoms, amounts
}

// TODO: LEAVING IT HERE FOR NOW (22/02/24 21:02)
// - Make all the X_banker functions work as they should
// - Remove the identifier mapping logic from genstd.
// - Possibly move all std types to std/types. This avoids any clashes with gno precompilation.
// - Make precompilation understand that bodyless function -> native function.
// - Add whether the *gno.Machine parameter is present to the native function data.
// When you get back on track for the transpiler:
// - Make StaticCheck work
// - Add a way to recursively precompile dependencies
// - Work until gno transpile --gobuild can create fully buildable go code!

func X_bankerSendCoins(m *gno.Machine, bt uint8, from, to string, denoms []string, amounts []int64) {
	// bt != BankerTypeReadonly (checked in gno)
	if bt == btOrigSend {
		if from != osb.pkgAddr {
			panic(fmt.Sprintf(
				"OrigSendBanker can only send from the realm package address %q, but got %q",
				osb.pkgAddr, from))
		}
		spent := (*osb.origSendSpent).Add(amt)
		if !osb.origSend.IsAllGTE(spent) {
			panic(fmt.Sprintf(
				`cannot send "%v", limit "%v" exceeded with "%v" already spent`,
				amt, osb.origSend, *osb.origSendSpent))
		}
		osb.banker.SendCoins(from, to, amt)
		*osb.origSendSpent = spent
	}
}

func X_bankerTotalCoin(m *gno.Machine, bt uint8, denom string) int64 {
	return m.Context.(ExecContext).Banker.TotalCoin(denom)
}
func X_bankerIssueCoin(m *gno.Machine, bt uint8, addr string, denom string, amount string)
func X_bankerRemoveCoin(m *gno.Machine, bt uint8, addr string, denom string, amount string)

func (osb OrigSendBanker) IssueCoin(addr crypto.Bech32Address, denom string, amount int64) {
	panic("OrigSendBanker cannot issue coins")
}

func (osb OrigSendBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amount int64) {
	panic("OrigSendBanker cannot remove coins")
}

//----------------------------------------
// RealmSendBanker

type RealmSendBanker struct {
	banker  Banker
	pkgAddr crypto.Bech32Address
}

func NewRealmSendBanker(banker Banker, pkgAddr crypto.Bech32Address) RealmSendBanker {
	return RealmSendBanker{
		banker:  banker,
		pkgAddr: pkgAddr,
	}
}

func (rsb RealmSendBanker) GetCoins(addr crypto.Bech32Address) (dst std.Coins) {
	return rsb.banker.GetCoins(addr)
}

func (rsb RealmSendBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins) {
	if from != rsb.pkgAddr {
		panic(fmt.Sprintf(
			"RealmSendBanker can only send from the realm package address %q, but got %q",
			rsb.pkgAddr, from))
	}
	rsb.banker.SendCoins(from, to, amt)
}

func (rsb RealmSendBanker) TotalCoin(denom string) int64 {
	return rsb.banker.TotalCoin(denom)
}

func (rsb RealmSendBanker) IssueCoin(addr crypto.Bech32Address, denom string, amount int64) {
	panic("RealmSendBanker cannot issue coins")
}

func (rsb RealmSendBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amount int64) {
	panic("RealmSendBanker cannot remove coins")
}
