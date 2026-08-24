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
	// Tolerate a genesis that predates PreprocessGasPerByte (an old-binary
	// export omits the field, decoding it as zero); applyLegacyDefaults fills
	// it so validation matches GetParams' runtime behavior. InitGenesis stores
	// the defaulted value.
	err := gs.Params.applyLegacyDefaults().Validate()
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
	gs.Params = gs.Params.applyLegacyDefaults()
	if err := ValidateGenesis(gs); err != nil {
		panic(err)
	}
	if err := vm.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}
	// Warn, loudly, on the one configuration that cannot be recovered from.
	//
	// run_submitters gates MsgRun and fails closed, and govdao proposal
	// creation is MsgRun-only (a ProposalRequest carries a func value, which
	// MsgCall cannot marshal). So an empty list means no proposal can ever be
	// created — including the proposal that would populate the list. Unlike the
	// other allowlists, whose empty state merely disables a capability while
	// governance keeps working, this one disables the means of repair.
	//
	// Deliberately not a Validate error: the field's zero value IS empty, so
	// rejecting it would make DefaultParams() invalid and break every chain
	// that does not use MsgRun at all. A warning is the strongest thing that
	// can be said here without that. It fires only where it should — test
	// genesis, gnodev and any seeded chain populate the list, so a quiet boot
	// means it was set.
	if len(gs.Params.RunSubmitters) == 0 {
		ctx.Logger().Error(
			"vm: run_submitters is empty, so no address may send MsgRun. " +
				"Governance proposal creation requires MsgRun, so this cannot be " +
				"fixed on-chain: set vm.run_submitters in genesis " +
				"(gnogenesis params set vm.run_submitters <addr>,...) before starting a chain " +
				"that expects governance to work.")
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
