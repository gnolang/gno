// Package chain contains core types and their helper functions to deal with the
// blockchain.
package chain

import (
	iruntime "internal/runtime"
)

// Address represents a bech32-encoded blockchain address.
type Address string

func (a Address) IsValid() bool {
	// TODO: make stdlib crypto/bech32
	_, _, ok := DecodeBech32(string(a))
	return ok
}

func NewAddressFromImportPath(importPath string) {
	// TODO: find better name
	panic("not implemented")
}

// A Realm represents an on-chain entity which has an address. This is generally
// an Externally Owned Account (ie. a "user") or a code realm, which also has an
// associated import path.
type Realm struct {
	address    Address
	importPath string
}

func (r Realm) Address() Address {
	return r.addr
}

func (r Realm) ImportPath() string {
	return r.importPath
}

func (r Realm) IsUser() bool {
	return r.importPath == ""
}

// NewCodeRealm creates a new realm, representing a code realm with a published
// import path, existing on-chain.
func NewCodeRealm(importPath string) Realm {
	if !iruntime.IsTesting() && !iruntime.IsRuntime() {
		panic("realms may only be created in testing or by the runtime package")
	}
	return Realm{
		address:    NewAddressFromImportPath(importPath),
		importPath: importPath,
	}
}

// NewUserRealm creates a new realm, representing a regular user.
func NewUserRealm(address Address) Realm {
	if !iruntime.IsTesting() && !iruntime.IsRuntime() {
		panic("realms may only be created in testing or by the runtime package")
	}
	return Realm{
		address: address,
	}
}

// Banker is an interface to access the blockchain's native banker.
type Banker interface {
	// Balance returns the balance in the bank for the given address, as a list
	// of coins.
	Balance(addr Address) (dst Coins)

	// Supply returns the circulating supply of the given denomination.
	Supply(denom string) int64

	// Send sends the given Coins between the from and to addresses.
	Send(from, to Address, amt Coins)

	// Issue issues the coin with the given denomination
	Issue(addr Address, denom string, amount int64)
	Burn(addr Address, denom string, amount int64)
}
