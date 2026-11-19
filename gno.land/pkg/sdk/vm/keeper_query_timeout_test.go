package vm

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const spinRealmPath = "gno.land/r/spin"

const spinRealmSource = `package spin

type T struct{ N int }

var List = []*T{&T{1}, &T{2}}

func GetList() []*T { return List }

// Spin runs a tight loop of n iterations.
func Spin(n int) int {
	s := 0
	for i := 0; i < n; i++ {
		s += i
	}
	return s
}

// SpinRecover is Spin with a recover around every iteration.
func SpinRecover(n int) int {
	s := 0
	for i := 0; i < n; i++ {
		func() {
			defer func() { recover() }()
			s += i
		}()
	}
	return s
}

type slowErr struct{}

func (slowErr) Error() string {
	for {
	}
}

// SlowErr returns an error whose Error never returns.
func SlowErr() error { return slowErr{} }

// PanicSlowErr panics with an error whose Error never returns.
func PanicSlowErr() int { panic(slowErr{}) }

// Mimic panics with the text of a stop by the deadline.
func Mimic() int { panic("deadline exceeded in location: CPUCycles") }

func Render(path string) string {
	Spin(1 << 62)
	return ""
}
`

// queryWithTimeout runs one vm query through the handler with a deadline armed
// the way BaseApp.Query arms it. A timeout <= 0 starts it already expired.
func queryWithTimeout(env testEnv, timeout time.Duration, path, data string) abci.ResponseQuery {
	expired := new(atomic.Bool)
	if timeout <= 0 {
		expired.Store(true)
	} else {
		t := time.AfterFunc(timeout, func() { expired.Store(true) })
		defer t.Stop()
	}
	return env.vmh.Query(sdk.WithQueryDeadline(queryCtx(env), expired), abci.RequestQuery{Path: path, Data: []byte(data)})
}

// TestQueryTimeout checks that a charged loop in qeval, qrender and qeval_json
// stops at the deadline, long before its gas budget, with QueryTimeoutError.
// No recover holds it off: not one in the realm, not the one qeval_json runs
// around a result's Error method, nor the one around a panic value's Error
// method when the panic is printed.
func TestQueryTimeout(t *testing.T) {
	env := setupSizeGuardRealm(t, "spinaddr", spinRealmPath, spinRealmSource)
	const timeout = 20 * time.Millisecond

	for _, ep := range []struct{ path, data string }{
		{"vm/qeval", spinRealmPath + ".Spin(1000000000000)"},
		{"vm/qeval", spinRealmPath + ".SpinRecover(1000000000000)"},
		{"vm/qeval_json", spinRealmPath + ".Spin(1000000000000)"},
		{"vm/qeval_json", spinRealmPath + ".SpinRecover(1000000000000)"},
		{"vm/qeval_json", spinRealmPath + ".SlowErr()"},
		{"vm/qeval", spinRealmPath + ".PanicSlowErr()"},
		{"vm/qrender", spinRealmPath + ":"},
	} {
		start := time.Now()
		res := queryWithTimeout(env, timeout, ep.path, ep.data)
		require.IsType(t, std.QueryTimeoutError{}, res.Error, "%s %s: %s", ep.path, ep.data, res.Log)
		// Generous: the point is it stops near the timeout, not at 3e9 gas.
		assert.Less(t, time.Since(start), timeout+2*time.Second, ep.data)
	}
}

// TestQueryTimeoutNotOtherErrors checks that a stop by gas stays out of gas
// with a deadline armed, and that a realm panic reading like the timeout is not
// reported as one.
func TestQueryTimeoutNotOtherErrors(t *testing.T) {
	env := setupSizeGuardRealm(t, "spinaddr", spinRealmPath, spinRealmSource)

	env.vmk.QueryMaxGas = 1_000_000
	res := queryWithTimeout(env, time.Hour, "vm/qeval", spinRealmPath+".Spin(1000000000000)")
	require.Error(t, res.Error)
	require.NotEqual(t, std.QueryTimeoutError{}, res.Error, res.Log)
	assert.Contains(t, res.Log, "OutOfGasError")
	env.vmk.QueryMaxGas = DefaultQueryMaxGas

	res = queryWithTimeout(env, time.Hour, "vm/qeval", spinRealmPath+".Mimic()")
	require.Error(t, res.Error)
	require.NotEqual(t, std.QueryTimeoutError{}, res.Error, res.Log)
	assert.Contains(t, res.Log, "deadline exceeded")

	// No deadline armed: the query runs to its end.
	out, err := env.vmk.QueryEval(queryCtx(env), spinRealmPath, "Spin(1000)")
	require.NoError(t, err)
	assert.Equal(t, "(499500 int)", out)
}

// TestVmHandlerQuery_DeadlineEveryEndpoint runs every vm query endpoint past
// its deadline. Each must answer QueryTimeoutError: several stop before any
// machine exists, where only a doRecoverQueryNoMachine defer reports it.
func TestVmHandlerQuery_DeadlineEveryEndpoint(t *testing.T) {
	env := setupSizeGuardRealm(t, "spinaddr", spinRealmPath, spinRealmSource)
	for _, ep := range everyQueryEndpoint(t, env, spinRealmPath) {
		var res abci.ResponseQuery
		require.NotPanics(t, func() { res = queryWithTimeout(env, 0, ep.path, ep.data) }, ep.path)
		require.IsType(t, std.QueryTimeoutError{}, res.Error, "%s: %s", ep.path, res.Log)
	}
}
