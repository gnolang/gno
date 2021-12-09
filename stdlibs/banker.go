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
	GetCoins(addr crypto.Address, dst *std.Coins)
	SendCoins(from, to crypto.Address, amt std.Coins)
	TotalCoin(denom string) int64
	IssueCoin(addr crypto.Address, denom string, amount int64)
	RemoveCoin(addr crypto.Address, denom string, amount int64)
}

// Used in std.GetBanker(options).
// Also available as Gno in stdlibs/std/banker.go
type BankerType uint8

// Also available as Gno in stdlibs/std/banker.go
const (
	// Can only read state.
	BankerTypeReadonly BankerType = iota
	// Can only send from tx send.
	BankerTypeTxSend
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

func (rb ReadonlyBanker) GetCoins(addr crypto.Address, dst *std.Coins) {
	rb.banker.GetCoins(addr, dst)
}
func (rb ReadonlyBanker) SendCoins(from, to crypto.Address, amt std.Coins) {
	panic("ReadonlyBanker cannot send coins")
}
func (rb ReadonlyBanker) TotalCoin(denom string) int64 {
	return rb.banker.TotalCoin(denom)
}
func (rb ReadonlyBanker) IssueCoin(addr crypto.Address, denom string, amount int64) {
	panic("ReadonlyBanker cannot issue coins")
}
func (rb ReadonlyBanker) RemoveCoin(addr crypto.Address, denom string, amount int64) {
	panic("ReadonlyBanker cannot remove coins")
}

//----------------------------------------
// TxSendBanker

type TxSendBanker struct {
	banker      Banker
	pkgAddr     crypto.Address
	txSend      std.Coins
	txSendSpent *std.Coins
}

func NewTxSendBanker(banker Banker, pkgAddr crypto.Address, txSend std.Coins, txSendSpent *std.Coins) TxSendBanker {
	if txSendSpent == nil {
		panic("txSendSpent cannot be nil")
	}
	return TxSendBanker{
		banker:      banker,
		pkgAddr:     pkgAddr,
		txSend:      txSend,
		txSendSpent: txSendSpent,
	}
}

func (tsb TxSendBanker) GetCoins(addr crypto.Address, dst *std.Coins) {
	tsb.banker.GetCoins(addr, dst)
}
func (tsb TxSendBanker) SendCoins(from, to crypto.Address, amt std.Coins) {
	if from != tsb.pkgAddr {
		panic("TxSendBanker can only send from the realm package address")
	}
	spent := (*tsb.txSendSpent).Add(amt)
	if !tsb.txSend.IsAllGTE(spent) {
		panic(fmt.Sprintf(
			"cannot send %v, limit %v exceeded with %v already spent",
			amt, tsb.txSend, *tsb.txSendSpent))
	}
	tsb.banker.SendCoins(from, to, amt)
	*tsb.txSendSpent = spent
}
func (tsb TxSendBanker) TotalCoin(denom string) int64 {
	return tsb.banker.TotalCoin(denom)
}
func (tsb TxSendBanker) IssueCoin(addr crypto.Address, denom string, amount int64) {
	panic("TxSendBanker cannot issue coins")
}
func (tsb TxSendBanker) RemoveCoin(addr crypto.Address, denom string, amount int64) {
	panic("TxSendBanker cannot remove coins")
}

//----------------------------------------
// RealmSendBanker

type RealmSendBanker struct {
	banker  Banker
	pkgAddr crypto.Address
}

func NewRealmSendBanker(banker Banker, pkgAddr crypto.Address) RealmSendBanker {
	return RealmSendBanker{
		banker:  banker,
		pkgAddr: pkgAddr,
	}
}

func (rsb RealmSendBanker) GetCoins(addr crypto.Address, dst *std.Coins) {
	rsb.banker.GetCoins(addr, dst)
}
func (rsb RealmSendBanker) SendCoins(from, to crypto.Address, amt std.Coins) {
	if from != rsb.pkgAddr {
		panic("RealmSendBanker can only send from the realm package address")
	}
	rsb.banker.SendCoins(from, to, amt)
}
func (rsb RealmSendBanker) TotalCoin(denom string) int64 {
	return rsb.banker.TotalCoin(denom)
}
func (rsb RealmSendBanker) IssueCoin(addr crypto.Address, denom string, amount int64) {
	panic("RealmSendBanker cannot issue coins")
}
func (rsb RealmSendBanker) RemoveCoin(addr crypto.Address, denom string, amount int64) {
	panic("RealmSendBanker cannot remove coins")
}
