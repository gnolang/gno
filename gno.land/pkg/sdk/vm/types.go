package vm

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// AccountKeeperI is the limited interface only needed for VM.
type AccountKeeperI interface {
	GetAccount(ctx sdk.Context, addr crypto.Address) std.Account
}

// BankKeeperI is the limited interface only needed for VM.
type BankKeeperI interface {
	GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins
	SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error)
	AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error)
	RestrictedDenoms(ctx sdk.Context) []string
}

// ParamsKeeperI is the limited interface only needed for VM.
type ParamsKeeperI interface {
	params.ParamsKeeperI

	IsRegistered(moduleName string) bool
	GetRegisteredKeeper(moduleName string) params.ParamfulKeeper
}

// Public facing function signatures.
// See convertArgToGno() for supported types.
type FunctionSignature struct {
	FuncName string
	Params   []NamedType
	Results  []NamedType
}

type NamedType struct {
	Name  string
	Type  string
	Value string
}

type FunctionSignatures []FunctionSignature

func (fsigs FunctionSignatures) JSON() string {
	bz := amino.MustMarshalJSON(fsigs)
	return string(bz)
}
