package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// InitGenesis - Init store state from genesis data
func (ak AccountKeeper) InitGenesis(ctx sdk.Context, data GenesisState) {
	if amino.DeepEqual(data, GenesisState{}) {
		if err := ak.SetParams(ctx, DefaultParams()); err != nil {
			panic(err)
		}
		return
	}

	if err := ValidateGenesis(data); err != nil {
		panic(err)
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
