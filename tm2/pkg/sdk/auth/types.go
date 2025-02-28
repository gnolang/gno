package auth

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// AccountKeeper manages access to accounts.
type AccountKeeperI interface {
	NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) std.Account
	GetAccount(ctx sdk.Context, addr crypto.Address) std.Account
	GetAllAccounts(ctx sdk.Context) []std.Account
	SetAccount(ctx sdk.Context, acc std.Account)
	IterateAccounts(ctx sdk.Context, process func(std.Account) bool)
	InitGenesis(ctx sdk.Context, data GenesisState)
	GetParams(ctx sdk.Context) Params
}

var _ AccountKeeperI = AccountKeeper{}

// Limited interface only needed for auth.
type BankKeeperI interface {
	SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
}

type GasPriceKeeperI interface {
	LastGasPrice(ctx sdk.Context) std.GasPrice
	SetGasPrice(ctx sdk.Context, gp std.GasPrice)
	UpdateGasPrice(ctx sdk.Context)
}

var _ GasPriceKeeperI = GasPriceKeeper{}
