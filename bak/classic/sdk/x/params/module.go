package params

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/client/context"
	"github.com/tendermint/classic/sdk/types/module"
	"github.com/tendermint/classic/sdk/x/params/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

const moduleName = "params"

// app module basics object
type AppModuleBasic struct{}

// module name
func (AppModuleBasic) Name() string {
	return moduleName
}

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage { return nil }

// module validate genesis
func (AppModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }
