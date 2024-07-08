package sdk

import (
	"fmt"
	"regexp"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (app *BaseApp) Check(tx Tx) (result Result) {
	return app.runTx(RunTxModeCheck, nil, tx)
}

func (app *BaseApp) Simulate(txBytes []byte, tx Tx) (result Result) {
	return app.runTx(RunTxModeSimulate, txBytes, tx)
}

func (app *BaseApp) Deliver(tx Tx) (result Result) {
	return app.runTx(RunTxModeDeliver, nil, tx)
}

// Context with current {check, deliver}State of the app
// used by tests
func (app *BaseApp) NewContext(mode RunTxMode, header abci.Header) Context {
	if mode == RunTxModeCheck {
		return NewContext(mode, app.checkState.ms, header, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}

	return NewContext(mode, app.deliverState.ms, header, app.logger)
}

// TODO: replace with abci.ABCIErrorOrStringError().
func ABCIError(err error) abci.Error {
	return abci.ABCIErrorOrStringError(err)
}

func ABCIResultFromError(err error) (res Result) {
	res.Error = ABCIError(err)
	res.Log = fmt.Sprintf("%#v", err)
	return
}

func ABCIResponseQueryFromError(err error) (res abci.ResponseQuery) {
	res.Error = ABCIError(err)
	res.Log = fmt.Sprintf("%#v", err)
	return
}
