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
	// A note, not a warning, and deliberately at Info.
	//
	// An empty run_submitters means the MsgRun allowlist is OFF — anyone may
	// send it — which is the behaviour that predates the param and so is not a
	// misconfiguration. What is worth one line in the log is that a chain
	// intending to gate MsgRun has not done so yet, because nothing else about
	// the boot distinguishes "left open on purpose" from "forgot".
	//
	// Never a Validate error: the field's zero value IS empty, so rejecting it
	// would make DefaultParams() invalid.
	if len(gs.Params.RunSubmitters) == 0 {
		ctx.Logger().Info(
			"vm: run_submitters is empty, so any address may send MsgRun. " +
				"To restrict it, list the permitted addresses " +
				"(gnogenesis params set vm.run_submitters <addr>,... at genesis, " +
				"or a governance proposal on a running chain). " +
				"Keep at least one address that can create governance proposals: " +
				"proposal creation is MsgRun-only.")
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
