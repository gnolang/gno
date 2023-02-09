package stdlibs

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

// This has the same interface as stdlibs/std.Banker.
// The native implementation of Banker (wrapped by any
// wrappers for limiting functionality as necessary)
// becomes available in Gno that implements
// stdlibs/std.Banker.
type Banker interface {
	GetCoins(addr crypto.Bech32Address) (dst std.Coins)
	SendCoins(from, to crypto.Bech32Address, amt std.Coins)
	TotalCoin(denom string) int64
	IssueCoin(addr crypto.Bech32Address, denom string, amount int64)
	RemoveCoin(addr crypto.Bech32Address, denom string, amount int64)
}

// Used in std.GetBanker(options).
// Also available as Gno in stdlibs/std/banker.go
type BankerType uint8

// Also available as Gno in stdlibs/std/banker.go
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
// ReadonlyBanker

type ReadonlyBanker struct {
	banker Banker
}

func NewReadonlyBanker(banker Banker) ReadonlyBanker {
	return ReadonlyBanker{banker}
}

func (rb ReadonlyBanker) GetCoins(addr crypto.Bech32Address) (dst std.Coins) {
	return rb.banker.GetCoins(addr)
}

func (rb ReadonlyBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins) {
	panic("ReadonlyBanker cannot send coins")
}

func (rb ReadonlyBanker) TotalCoin(denom string) int64 {
	return rb.banker.TotalCoin(denom)
}

func (rb ReadonlyBanker) IssueCoin(addr crypto.Bech32Address, denom string, amount int64) {
	panic("ReadonlyBanker cannot issue coins")
}

func (rb ReadonlyBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amount int64) {
	panic("ReadonlyBanker cannot remove coins")
}

//----------------------------------------
// OrigSendBanker

type OrigSendBanker struct {
	banker        Banker
	pkgAddr       crypto.Bech32Address
	origSend      std.Coins
	origSendSpent *std.Coins
}

func NewOrigSendBanker(banker Banker, pkgAddr crypto.Bech32Address, origSend std.Coins, origSendSpent *std.Coins) OrigSendBanker {
	if origSendSpent == nil {
		panic("origSendSpent cannot be nil")
	}
	return OrigSendBanker{
		banker:        banker,
		pkgAddr:       pkgAddr,
		origSend:      origSend,
		origSendSpent: origSendSpent,
	}
}

func (osb OrigSendBanker) GetCoins(addr crypto.Bech32Address) (dst std.Coins) {
	return osb.banker.GetCoins(addr)
}

func (osb OrigSendBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins) {
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

func (osb OrigSendBanker) TotalCoin(denom string) int64 {
	return osb.banker.TotalCoin(denom)
}

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
