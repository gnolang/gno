package sdk

import (
	"fmt"
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (app *BaseApp) Check(tx Tx) (result Result) {
	txBytes, err := amino.Marshal(tx)
	if err != nil {
		return ABCIResultFromError(std.ErrTxDecode(err.Error()))
	}
	ctx := app.getContextForTx(RunTxModeCheck, nil)
	return app.runTx(ctx, txBytes)
}

func (app *BaseApp) Simulate(txBytes []byte) (result Result) {
	// Read header from the atomic snapshot — safe for concurrent access.
	header := app.getLastBlockHeader()
	if header == nil || header.GetHeight() < 1 {
		// Before first commit (e.g., during InitChain or tests),
		// fall back to checkState which is safe in single-threaded context.
		ctx := app.getContextForTx(RunTxModeSimulate, txBytes)
		return app.runTx(ctx, txBytes)
	}

	height := header.GetHeight()

	// Load an immutable snapshot of committed state at the given height.
	// This is safe for concurrent access — IAVL versions are copy-on-write.
	cacheMS, err := app.cms.MultiImmutableCacheWrapWithVersion(height)
	if err != nil {
		return ABCIResultFromError(
			std.ErrInternal(fmt.Sprintf("failed to load state for simulate at height %d: %s", height, err)),
		)
	}

	ctx := NewContext(RunTxModeSimulate, cacheMS, header, app.logger).
		WithTxBytes(txBytes).
		WithMinGasPrices(app.minGasPrices).
		WithConsensusParams(app.consensusParams)

	return app.runTx(ctx, txBytes)
}

func (app *BaseApp) Deliver(tx Tx, ctxFns ...ContextFn) (result Result) {
	txBytes, err := amino.Marshal(tx)
	if err != nil {
		return ABCIResultFromError(std.ErrTxDecode(err.Error()))
	}
	ctx := app.getContextForTx(RunTxModeDeliver, nil)

	for _, ctxFn := range ctxFns {
		if ctxFn == nil {
			continue
		}

		ctx = ctxFn(ctx)
	}

	return app.runTx(ctx, txBytes)
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
