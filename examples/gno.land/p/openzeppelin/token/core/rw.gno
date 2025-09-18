package core

import (
	"std"

	"gno.land/p/onbloc/uint256"
	"gno.land/p/openzeppelin/storage"
)

// This file implements the keyed storage access for core token.

// TODO: Support iterated reads later if needed.
// TODO: Implement allowances later - it is a must for first implementation!

func IDKey() string {
	return "core/id"
}

func GetID(state storage.State) string {
	return getString(IDKey(), state)
}

func SetID(state storage.State, id string) {
	state.Set(IDKey(), id)
}

func CreatedKey() string {
	return "core/created"
}

func IsCreated(state storage.State) bool {
	v, ok := state.Get(CreatedKey())
	if !ok {
		return false
	}
	return v.(bool)
}

func SetCreated(state storage.State) {
	state.Set(CreatedKey(), true)
}

func BalanceKey(addr std.Address) string {
	return "core/balance/" + addr.String()
}

func GetBalance(state storage.State, addr std.Address) *uint256.Uint {
	v, ok := state.Get(BalanceKey(addr))
	if !ok {
		return uint256.NewUint(0)
	}
	return v.(*uint256.Uint)
}

func SetBalance(state storage.State, addr std.Address, val *uint256.Uint) {
	state.Set(BalanceKey(addr), val)
}

func NameKey() string {
	return "core/name"
}

func GetName(state storage.State) string {
	return getString(NameKey(), state)
}

func SetName(state storage.State, name string) {
	state.Set(NameKey(), name)
}

func SymbolKey() string {
	return "core/symbol"
}

func GetSymbol(state storage.State) string {
	return getString(SymbolKey(), state)
}

func SetSymbol(state storage.State, name string) {
	state.Set(SymbolKey(), name)
}

// TODO: Generalize these kind of helpers later as utils to speed up development.
func getString(key string, state storage.State) string {
	v, ok := state.Get(key)
	if !ok {
		return ""
	}
	return v.(string)
}
