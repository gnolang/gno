package sdk

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// spinDeadline charges a deadline meter built from ctx, the way the VM keeper
// builds a query's, until it panics.
func spinDeadline(ctx Context) {
	meter := store.NewDeadlineGasMeter(1<<60, QueryDeadline(ctx))
	for {
		meter.ConsumeGas(1, "spin")
	}
}

// TestBaseAppQueryTimeout checks that a custom-route query stopped by the
// deadline answers QueryTimeoutError, whether its handler panics or returns
// the error, and that the mapping goes by type: an error that only reads like
// a timeout, or that came after the deadline for another reason, keeps its own.
func TestBaseAppQueryTimeout(t *testing.T) {
	t.Parallel()

	mimic := store.DeadlineExceededError{Descriptor: "spin"}.Error()
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute("q", testHandler{query: func(ctx Context, req abci.RequestQuery) (res abci.ResponseQuery) {
			switch req.Path {
			case "q/panic":
				spinDeadline(ctx)
			case "q/error":
				// The keeper's way: recover the meter's panic into an error.
				defer func() {
					err := fmt.Errorf("wrapped: %w", recover().(error))
					res = ABCIResponseQueryFromError(err)
				}()
				spinDeadline(ctx)
			case "q/late":
				for !QueryDeadline(ctx).Load() {
					time.Sleep(time.Millisecond)
				}
				return ABCIResponseQueryFromError(std.ErrUnknownRequest("too large"))
			case "q/mimic-error":
				return ABCIResponseQueryFromError(errors.New(mimic))
			case "q/mimic-panic":
				panic(mimic)
			}
			return
		}})
	}
	app := setupBaseApp(t, routerOpt)
	app.SetQueryTimeout(5 * time.Millisecond)
	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})
	app.Commit()

	for _, path := range []string{"q/panic", "q/error"} {
		res := app.Query(abci.RequestQuery{Path: path})
		require.IsType(t, std.QueryTimeoutError{}, res.Error, "%s: %s", path, res.Log)
	}
	for path, want := range map[string]abci.Error{
		"q/late":        std.UnknownRequestError{},
		"q/mimic-error": abci.StringError(""),
		"q/mimic-panic": std.InternalError{},
	} {
		res := app.Query(abci.RequestQuery{Path: path})
		require.IsType(t, want, res.Error, "%s: %s", path, res.Log)
	}
}

// TestBaseAppQueryDeadlineOnlyCustom checks which paths carry the deadline
// flag: a custom-route query when a timeout is set, and nothing else. A
// simulate that runs well past the timeout still succeeds and reports all the
// gas it used, and CheckTx and DeliverTx never see the flag either.
func TestBaseAppQueryDeadlineOnlyCustom(t *testing.T) {
	t.Parallel()

	const timeout = 5 * time.Millisecond
	const gas = 1000
	// Where the handlers ran, and whether the flag was there.
	var seen []string
	record := func(where string, ctx Context) {
		seen = append(seen, fmt.Sprintf("%s:%v", where, QueryDeadline(ctx) != nil))
	}
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx Context, tx Tx, simulate bool) (Context, Result, bool) {
			record(fmt.Sprintf("ante/%d", ctx.Mode()), ctx)
			return ctx.WithGasMeter(store.NewGasMeter(1 << 30)), Result{}, false
		})
	}
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(routeMsgCounter, testHandler{
			process: func(ctx Context, msg Msg) Result {
				record(fmt.Sprintf("msg/%d", ctx.Mode()), ctx)
				// Outlive the timeout, then charge both the tx meter and a
				// meter built the way a query's is: neither may stop.
				time.Sleep(4 * timeout)
				ctx.GasMeter().ConsumeGas(gas, "work")
				store.NewDeadlineGasMeter(1<<30, QueryDeadline(ctx)).ConsumeGas(1, "work")
				return Result{}
			},
			query: func(ctx Context, req abci.RequestQuery) abci.ResponseQuery {
				record("query", ctx)
				return abci.ResponseQuery{}
			},
		})
	}
	app := setupBaseApp(t, anteOpt, routerOpt)
	app.SetQueryTimeout(timeout)
	app.InitChain(abci.RequestInitChain{ChainID: "test-chain"})
	app.Commit()

	txBytes, err := amino.Marshal(newTxCounter(0, 0))
	require.NoError(t, err)

	res := app.Query(abci.RequestQuery{Path: ".app/simulate", Data: txBytes})
	require.True(t, res.IsOK(), res.Log)
	var simRes Result
	require.NoError(t, amino.Unmarshal(res.Value, &simRes))
	require.True(t, simRes.IsOK(), simRes.Log)
	require.GreaterOrEqual(t, simRes.GasUsed, int64(gas))

	require.True(t, app.CheckTx(abci.RequestCheckTx{Tx: txBytes}).IsOK())
	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: "test-chain", Height: 2}})
	dres := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, dres.IsOK(), dres.Log)

	require.True(t, app.Query(abci.RequestQuery{Path: routeMsgCounter + "/x"}).IsOK())
	app.SetQueryTimeout(0)
	require.True(t, app.Query(abci.RequestQuery{Path: routeMsgCounter + "/x"}).IsOK())

	sim, check, deliver := RunTxModeSimulate, RunTxModeCheck, RunTxModeDeliver
	require.Equal(t, []string{
		fmt.Sprintf("ante/%d:false", sim), fmt.Sprintf("msg/%d:false", sim),
		fmt.Sprintf("ante/%d:false", check),
		fmt.Sprintf("ante/%d:false", deliver), fmt.Sprintf("msg/%d:false", deliver),
		"query:true", "query:false",
	}, seen)
}
