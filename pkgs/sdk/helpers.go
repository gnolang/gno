package sdk

import (
	"fmt"
	"regexp"

	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/errors"
)

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// nolint - Mostly for testing
func (app *BaseApp) Check(tx Tx) (result Result) {
	return app.runTx(RunTxModeCheck, nil, tx)
}

// nolint - full tx execution
func (app *BaseApp) Simulate(txBytes []byte, tx Tx) (result Result) {
	return app.runTx(RunTxModeSimulate, txBytes, tx)
}

// nolint
func (app *BaseApp) Deliver(tx Tx) (result Result) {
	return app.runTx(RunTxModeDeliver, nil, tx)
}

// Context with current {check, deliver}State of the app
// used by tests
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) Context {
	if isCheckTx {
		return NewContext(app.checkState.ms, header, true, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}

	return NewContext(app.deliverState.ms, header, false, app.logger)
}

func ABCIError(err error) abci.Error {
	if err == nil {
		return nil
	}
	err = errors.Cause(err) // unwrap
	abcierr, ok := err.(abci.Error)
	if !ok {
		return abci.StringError(err.Error())
	} else {
		return abcierr
	}
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
