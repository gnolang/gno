package core

import (
	"std"
	"strings"

	"gno.land/p/onbloc/uint256"
	"gno.land/p/openzeppelin/storage"
)

func Create(tokenID, name, symbol string, minter std.Address, mintAmount *uint256.Uint, allowedPkgs ...string) storage.ExecFn {
	return func(state storage.State) any {
		if IsCreated(state) {
			panic("token already created")
		}

		storage.SetACL(state, allowedPkgs)

		SetName(state, name)
		SetSymbol(state, symbol)
		SetBalance(state, minter, mintAmount)
		SetID(state, tokenID)
		SetCreated(state)

		std.Emit(
			"TokenCreated",
			tokenID, name, symbol, minter.String(), mintAmount.String(),
			strings.Join(allowedPkgs, ","),
		)
		return tokenID
	}
}

func Transfer(from, to std.Address, amount *uint256.Uint) storage.ExecFn {
	return func(state storage.State) any {
		fromBalance := GetBalance(state, from)
		fromBalance.Sub(fromBalance, amount)
		SetBalance(state, from, fromBalance)
		toBalance := GetBalance(state, to)
		toBalance.Add(toBalance, amount)
		SetBalance(state, to, toBalance)
		std.Emit(
			"Transfer",
			"from", from.String(),
			"to", to.String(),
			"amount", amount.String(),
		)
		return nil
	}
}

func BalanceOf(addr std.Address) storage.ExecFn {
	return func(state storage.State) any {
		return GetBalance(state, addr)
	}
}

func Name() storage.ExecFn {
	return func(state storage.State) any {
		return GetName(state)
	}
}

func Symbol() storage.ExecFn {
	return func(state storage.State) any {
		return GetSymbol(state)
	}
}
