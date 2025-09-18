package v1

import (
	"encoding/hex"
	"std"
	"time"

	"gno.land/p/onbloc/uint256"
	pluglib "gno.land/p/openzeppelin/plugins"
	storagelib "gno.land/p/openzeppelin/storage"
	"gno.land/p/openzeppelin/token/core"
	"gno.land/p/samcrew/keccak256"
	"gno.land/r/openzeppelin/plugins"
	"gno.land/r/openzeppelin/storage"
)

// TODO: Allow module management after introducing user access control.

// TODO: Implement allowance or make it a plugin.

func caller() std.Address {
	return std.PreviousRealm().Address()
}

func randomish() string {
	// TODO: Improve.
	seed := []byte(std.OriginCaller().String() + time.Now().String())
	h := keccak256.Hash(seed)
	return hex.EncodeToString(h[:])
}

const PluginTransferHook = "TransferHook"

func Create(cur realm, name, symbol string, minter std.Address, mintAmount *uint256.Uint) string {
	tokenID := randomish()
	storage.Execute(
		cross, tokenID,
		core.Create(
			tokenID, name, symbol, minter, mintAmount,
			// Bootstrap storage with default allowed packages - to be managed later by the user.
			"gno.land/r/openzeppelin/fungible/v1",
			"gno.land/r/openzeppelin/plugins",
		),
	)

	// These plugins can be managed externally for a token by a privileged user.
	plugins.SetPlugins(cross, tokenID, []*pluglib.Plugin{
		{
			ID: PluginTransferHook,
			Func: func(state storagelib.State, args ...any) {
				symbol := args[0].(string)
				// Just emitting an event with a symbol to demonstrate the ability.
				std.Emit(PluginTransferHook, "symbol", symbol)
			},
		},
	})

	return tokenID
}

func Transfer(cur realm, tokenID string, to std.Address, amount *uint256.Uint) {
	storage.Execute(cross, tokenID, core.Transfer(caller(), to, amount))
	symbol := storage.Execute(cross, tokenID, core.Symbol()).(string)
	plugins.Run(cross, tokenID, PluginTransferHook, symbol)
}

func BalanceOf(cur realm, tokenID string, addr std.Address) *uint256.Uint {
	return storage.Execute(cross, tokenID, core.BalanceOf(addr)).(*uint256.Uint)
}

func Name(cur realm, tokenID string) string {
	return storage.Execute(cross, tokenID, core.Name()).(string)
}

func Symbol(cur realm, tokenID string) string {
	return storage.Execute(cross, tokenID, core.Symbol()).(string)
}
