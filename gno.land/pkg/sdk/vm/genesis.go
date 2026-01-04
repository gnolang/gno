package vm

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
)

// GenesisState - all state that must be provided at genesis
type GenesisState struct {
	Params      Params         `json:"params" yaml:"params"`
	RealmParams []params.Param `json:"realm_params" yaml:"realm_params"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params) GenesisState {
	return GenesisState{
		Params: params,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultParams())
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
// XXX refactor to .ValidateBasic() method.
func ValidateGenesis(gs GenesisState) error {
	if amino.DeepEqual(gs, GenesisState{}) {
		return fmt.Errorf("vm genesis state cannot be empty")
	}
	err := gs.Params.Validate()
	if err != nil {
		return err
	}
	// XXX validate RealmParams.
	// 1. all keys must be realm paths.
	// 2. all values must be supported types.
	return nil
}

// InitGenesis - Init store state from genesis data
func (vm *VMKeeper) InitGenesis(ctx sdk.Context, gs GenesisState) {
	if err := ValidateGenesis(gs); err != nil {
		panic(err)
	}
	if err := vm.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}
	// NOTE realm params should not have side effects so the order
	// shouldn't matter, but amino doesn't support maps (for determinism).
	for _, rp := range gs.RealmParams {
		vm.prmk.SetAny(ctx, "vm:"+rp.Key, rp.Value)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func (vm *VMKeeper) ExportGenesis(ctx sdk.Context) GenesisState {
	params := vm.GetParams(ctx)
	return NewGenesisState(params)
}
