package sdk

import (
	"fmt"
	"regexp"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (app *BaseApp) Check(tx Tx) (result Result) {
	ctx := app.getContextForTx(RunTxModeCheck, nil)

	return app.runTx(ctx, tx)
}

func (app *BaseApp) Simulate(txBytes []byte, tx Tx) (result Result) {
	ctx := app.getContextForTx(RunTxModeSimulate, txBytes)

	return app.runTx(ctx, tx)
}

func (app *BaseApp) Deliver(tx Tx, ctxFns ...ContextFn) (result Result) {
	ctx := app.getContextForTx(RunTxModeDeliver, nil)

	for _, ctxFn := range ctxFns {
		if ctxFn == nil {
			continue
		}

		ctx = ctxFn(ctx)
	}

	return app.runTx(ctx, tx)
}

// ContextFn is the custom execution context builder.
// It can be used to add custom metadata when replaying transactions
// during InitChainer or in the context of a unit test.
type ContextFn func(ctx Context) Context

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
