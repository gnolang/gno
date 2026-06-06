package sdk

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Log clipping limits — defense in depth backstop for adversarial
// content reaching ABCI Log. Per-line cap stops a single huge line
// (e.g. a wrapped error inlined into a stacktrace); line-count cap
// stops a tall stacktrace. Net cap: maxLogLines × (maxLogLineBytes
// + suffix) ≈ 17 KB.
const (
	maxLogLineBytes = 1024
	maxLogLines     = 16
)

// clipLog enforces two bounds on persisted Log content: per-line
// byte cap and total line-count cap. Truncated lines and elided
// tails get explicit markers. Fast-path returns input unchanged
// when it's small and contains no newlines.
func clipLog(s string) string {
	if len(s) <= maxLogLineBytes && !strings.ContainsRune(s, '\n') {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if len(line) > maxLogLineBytes {
			lines[i] = line[:maxLogLineBytes] + "...<truncated>"
		}
	}
	if len(lines) > maxLogLines {
		elided := len(lines) - maxLogLines
		lines = append(lines[:maxLogLines],
			fmt.Sprintf("... %d more lines elided", elided))
	}
	return strings.Join(lines, "\n")
}

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
	ctx := app.getContextForTx(RunTxModeSimulate, txBytes)
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
	res.Log = clipLog(fmt.Sprintf("%#v", err))
	return
}

func ABCIResponseQueryFromError(err error) (res abci.ResponseQuery) {
	res.Error = ABCIError(err)
	res.Log = clipLog(fmt.Sprintf("%#v", err))
	return
}
