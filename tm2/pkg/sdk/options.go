package sdk

import (
	"fmt"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// File for storing in-package BaseApp optional functions,
// for options that need access to non-exported fields of the BaseApp

// SetPruningOptions sets pruning options on the multistore associated with the app
func SetPruningOptions(opts store.PruningOptions) func(*BaseApp) {
	return func(bap *BaseApp) {
		sopts := bap.cms.GetStoreOptions()
		sopts.PruningOptions = opts
		bap.cms.SetStoreOptions(sopts)
	}
}

// SetMinGasPrices returns an option that sets the minimum gas prices on the app.
func SetMinGasPrices(gasPricesStr string) func(*BaseApp) {
	gasPrices, err := ParseGasPrices(gasPricesStr)
	if err != nil {
		panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
	}

	return func(bap *BaseApp) { bap.setMinGasPrices(gasPrices) }
}

func (app *BaseApp) SetName(name string) {
	if app.sealed {
		panic("SetName() on sealed BaseApp")
	}
	app.name = name
}

// SetAppVersion sets the application's version string.
func (app *BaseApp) SetAppVersion(v string) {
	if app.sealed {
		panic("SetAppVersion() on sealed BaseApp")
	}
	app.appVersion = v
}

func (app *BaseApp) SetDB(db dbm.DB) {
	if app.sealed {
		panic("SetDB() on sealed BaseApp")
	}
	app.db = db
}

func (app *BaseApp) SetCMS(cms store.CommitMultiStore) {
	if app.sealed {
		panic("SetEndBlocker() on sealed BaseApp")
	}
	app.cms = cms
}

func (app *BaseApp) SetInitChainer(initChainer InitChainer) {
	if app.sealed {
		panic("SetInitChainer() on sealed BaseApp")
	}
	app.initChainer = initChainer
}

func (app *BaseApp) SetBeginBlocker(beginBlocker BeginBlocker) {
	if app.sealed {
		panic("SetBeginBlocker() on sealed BaseApp")
	}
	app.beginBlocker = beginBlocker
}

func (app *BaseApp) SetEndBlocker(endBlocker EndBlocker) {
	if app.sealed {
		panic("SetEndBlocker() on sealed BaseApp")
	}
	app.endBlocker = endBlocker
}

func (app *BaseApp) SetAnteHandler(ah AnteHandler) {
	if app.sealed {
		panic("SetAnteHandler() on sealed BaseApp")
	}
	app.anteHandler = ah
}

func (app *BaseApp) SetBeginTxHook(endTx BeginTxHook) {
	if app.sealed {
		panic("SetBeginTxHook() on sealed BaseApp")
	}
	app.beginTxHook = endTx
}

func (app *BaseApp) SetEndTxHook(endTx EndTxHook) {
	if app.sealed {
		panic("SetEndTxHook() on sealed BaseApp")
	}
	app.endTxHook = endTx
}
