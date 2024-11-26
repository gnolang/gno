package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// GenesisState - all state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params) GenesisState {
	return GenesisState{params}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams())
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return data.Params.Validate()
}

// InitGenesis - Init store state from genesis data
func (ak AccountKeeper) InitGenesis(ctx sdk.Context, data GenesisState) {
	if amino.DeepEqual(data, GenesisState{}) {
		return
	}

	if err := ValidateGenesis(data); err != nil {
		panic(err)
	}

	// The unrestricted address must have been created as one of the genesis accounts.
	// Otherwise, we cannot verify the unrestricted address in the genesis state.

	for _, addr := range data.Params.UnrestrictedAddrs {
		acc := ak.GetAccount(ctx, addr)
		acc.SetUnrestricted(true)
		ak.SetAccount(ctx, acc)
	}

	if err := ak.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func (ak AccountKeeper) ExportGenesis(ctx sdk.Context) GenesisState {
	params := ak.GetParams(ctx)

	return NewGenesisState(params)
}
