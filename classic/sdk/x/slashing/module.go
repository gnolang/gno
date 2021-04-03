package slashing

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/classic/abci/types"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/client/context"
	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/types/module"
	"github.com/tendermint/classic/sdk/x/slashing/client/cli"
	"github.com/tendermint/classic/sdk/x/slashing/client/rest"
	"github.com/tendermint/classic/sdk/x/slashing/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module basics object
type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

// module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// module validate genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(StoreKey)
}

//___________________________
// app module
type AppModule struct {
	AppModuleBasic
	keeper        Keeper
	stakingKeeper types.StakingKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, stakingKeeper types.StakingKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		stakingKeeper:  stakingKeeper,
	}
}

// module name
func (AppModule) Name() string {
	return ModuleName
}

// register invariants
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// module message route name
func (AppModule) Route() string {
	return RouterKey
}

// module handler
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// module querier route name
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// module querier
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, am.stakingKeeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return ModuleCdc.MustMarshalJSON(gs)
}

// module begin-block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// module end-block
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
