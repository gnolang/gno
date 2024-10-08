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
}

type GasPriceKeeperI interface {
	LastGasPrice(ctx sdk.Context) std.GasPrice
}

var _ GasPriceKeeperI = GasPriceKeeper{}

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"params"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params) GenesisState {
	return GenesisState{params}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams())
}

// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return data.Params.Validate()
}
