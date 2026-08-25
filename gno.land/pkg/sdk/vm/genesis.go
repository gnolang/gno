package vm

import (
	"fmt"
	"strings"

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
	// Realm params are keyed <realm-path>:<name>, and InitGenesis writes them as
	// "vm:"+key. So a key whose first segment is not a realm path lands in the
	// vm module's OWN parameter namespace instead of a realm's.
	//
	// That matters because the realm-param loop runs after SetParams and
	// overwrites it. A genesis section [vm:p] with run_submitters.strings = [...]
	// becomes the key "p:run_submitters", writes vm:p:run_submitters, and
	// silently replaces the validated Params value -- while bypassing the scalar
	// [vm] allowlist, which admits only chain_domain and sysnames_pkgpath.
	//
	// A realm path always contains a "/" (gno.land/r/...); the vm's own
	// submodules do not, so requiring one separates the two namespaces exactly.
	//
	// XXX still open: values are not checked against a supported-type list.
	for _, rp := range gs.RealmParams {
		realm, name, found := strings.Cut(rp.Key, ":")
		if !found || realm == "" || name == "" {
			return fmt.Errorf(
				"invalid realm param key %q: want <realm-path>:<name>", rp.Key)
		}
		if !strings.Contains(realm, "/") {
			return fmt.Errorf(
				"invalid realm param key %q: %q is not a realm path, so this would "+
					"write the vm module's own parameter vm:%s rather than a realm's",
				rp.Key, realm, rp.Key)
		}
		// One colon exactly. A realm writing a param at runtime goes through
		// sys/params' prmkey, which panics on a name containing a colon, so a
		// key with two is one only genesis can produce. It is also ambiguous:
		// this splits on the first colon to find the realm, while
		// realmFromKey (params_deposit.go) splits on the last, so the two would
		// attribute the same key to different realms.
		if strings.Contains(name, ":") {
			return fmt.Errorf(
				"invalid realm param key %q: the name %q must not contain a colon",
				rp.Key, name)
		}
	}
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
