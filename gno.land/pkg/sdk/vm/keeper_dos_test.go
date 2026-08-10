package vm

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dosRealmSource returns functions that build large ephemeral values, used to
// reproduce the unmetered export-walk + JSON-marshal DoS on the JSON query
// endpoints and to verify the maxQueryExportBytes bound.
const dosRealmSource = `package dos

import "strings"

// Big returns an n-byte string. amino-marshaled unbounded this is the primary
// memory-exhaustion vector on qeval_json.
func Big(n int) string { return strings.Repeat("A", n) }

type Node struct {
	A, B, C, D int
	S          string
}

// Structs returns n structs; amino JSON of this is ~600 bytes/struct, so a
// modest metered allocation marshals to a much larger unmetered response.
func Structs(n int) []Node {
	s := make([]Node, n)
	for i := range s {
		s[i] = Node{A: i, S: "x"}
	}
	return s
}
`

// lowerExportBudget temporarily shrinks maxQueryExportBytes so the guard can be
// exercised with KB-sized values instead of allocating the real 10MB+ per case
// (a 12MB package var also pays a proportional storage deposit on deploy).
//
// It mutates a package global, so a test calling this — or any test whose
// result depends on the budget — must NOT call t.Parallel().
func lowerExportBudget(t *testing.T, n int64) {
	t.Helper()
	prev := maxQueryExportBytes
	maxQueryExportBytes = n
	t.Cleanup(func() { maxQueryExportBytes = prev })
}

// setupSizeGuardRealm deploys a single-file realm and commits it. MaxDeposit is
// always set: two of the callers store ~100KB, which the default does not cover.
func setupSizeGuardRealm(t *testing.T, seed, pkgPath, body string) testEnv {
	t.Helper()
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte(seed))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(100_000_000)))

	files := []*std.MemFile{
		// Sorts before gnomod.toml; AddPackage rejects unsorted mempackages.
		{Name: "a.gno", Body: body},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	msg := NewMsgAddPackage(addr, pkgPath, files)
	msg.MaxDeposit = std.MustParseCoins(ugnot.ValueString(50_000_000))
	require.NoError(t, env.vmk.AddPackage(ctx, msg))
	env.vmk.CommitGnoTransactionStore(ctx)
	return env
}

// TestQueryEval_SizeGuard is the regression test for the memory-exhaustion DoS,
// across both eval endpoints: a free query whose result would build a response
// past the budget must be rejected, not serviced. qeval_json marshals with
// amino, qeval renders with TypedValue.String(); both build the response
// outside the alloc meter, so both take the same budget.
func TestQueryEval_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/dos"
	lowerExportBudget(t, 50_000)
	env := setupSizeGuardRealm(t, "dosaddr", pkgPath, dosRealmSource)

	for _, ep := range []struct {
		name  string
		query func(ctx sdk.Context, pkgPath, expr string) (string, error)
	}{
		{"qeval", env.vmk.QueryEval},
		{"qeval_json", env.vmk.QueryEvalJSON},
	} {
		t.Run(ep.name, func(t *testing.T) {
			// Both exprs are comfortably under the eval alloc/gas ceiling, so
			// pre-fix they were serviced and serialized unmetered.
			for _, expr := range []string{"Big(200000)", "Structs(2000)"} {
				_, err := ep.query(env.ctx, pkgPath, expr)
				require.Error(t, err, expr)
				assert.ErrorIs(t, err, ExportSizeExceededError{}, expr)
			}
		})
	}

	// Small results are still serviced, in each endpoint's own format.
	t.Run("small_json_ok", func(t *testing.T) {
		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Big(10)")
		require.NoError(t, err)
		var got map[string]json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(res), &got))
		require.Contains(t, string(got["results"]), strings.Repeat("A", 10))

		res, err = env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Structs(3)")
		require.NoError(t, err)
		require.Contains(t, res, "results")
	})

	t.Run("small_render_ok", func(t *testing.T) {
		res, err := env.vmk.QueryEval(env.ctx, pkgPath, "Big(10)")
		require.NoError(t, err)
		assert.Contains(t, res, strings.Repeat("A", 10))
	})
}

// TestQueryObject_SizeGuard covers the shared exportObject path, i.e. both
// qobject_json and qobject_binary. The queried object is expanded inline, so an
// object that is large in itself — here a persisted byte array — must be
// rejected; the small object that references it still resolves.
func TestQueryObject_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/dosobj"
	lowerExportBudget(t, 20_000)

	// 100KB persisted byte array: charged at the base64 rate (~133KB), well
	// over the lowered budget once the array object itself is expanded.
	env := setupSizeGuardRealm(t, "dosobjaddr", pkgPath, `package dosobj

var Blob = make([]byte, 100000)

func GetBlob() []byte { return Blob }
`)

	// The slice value exports as a small RefValue to the backing array, so this
	// query is well under budget and hands us the array's ObjectID.
	res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetBlob()")
	require.NoError(t, err)
	oid := extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "no ObjectID in %s", res)

	// Resolving that ObjectID expands the array inline: rejected on both the
	// JSON and the binary endpoint, which share exportObject.
	_, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})

	_, err = env.vmk.QueryObjectBinary(env.ctx, oid)
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})
}

// TestQueryPkgJSON_SizeGuard verifies the qpkg_json path is bounded too: a
// package variable holding a large value must not force an unmetered export +
// marshal.
func TestQueryPkgJSON_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/dospkg"
	lowerExportBudget(t, 20_000)

	// 100KB var: over the lowered export budget, so the qpkg_json export walk
	// must reject it.
	env := setupSizeGuardRealm(t, "dospkgaddr", pkgPath, `package dospkg

import "strings"

var Big = strings.Repeat("Y", 100000)
`)

	_, err := env.vmk.QueryPkg(env.ctx, pkgPath)
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})
}

// qtypeAmpRealmSource is a struct DAG whose every field references the next
// named type. marshalTypeJSON re-expands DeclaredType.Base per reference with
// no seal, so this < 1KB of source expands to tens of MB of JSON (measured
// ~36MB for T0) — maxTypeDepth caps depth but not the per-level fanout.
const qtypeAmpRealmSource = `package amp

type T0 struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20, F21, F22, F23 T1
}
type T1 struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20, F21, F22, F23 T2
}
type T2 struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20, F21, F22, F23 T3
}
type T3 struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20, F21, F22, F23 T4
}
type T4 struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20, F21, F22, F23 T5
}
type T5 struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20, F21, F22, F23 int
}
func Render(path string) string { return "" }
`

// TestQueryType_SizeGuard covers the qtype_json path, which does not build an
// allocator and so is not bounded by the eval meters: a type whose hand-rolled
// JSON would blow past the budget must be refused, while an ordinary leaf type
// still resolves.
func TestQueryType_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/amp"
	lowerExportBudget(t, 50_000)
	env := setupSizeGuardRealm(t, "ampaddr", pkgPath, qtypeAmpRealmSource)

	// The wide DAG expands past the budget: refused, not serviced.
	_, err := env.vmk.QueryType(env.ctx, pkgPath+".T0")
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})

	// The leaf type is small and still resolves.
	res, err := env.vmk.QueryType(env.ctx, pkgPath+".T5")
	require.NoError(t, err)
	assert.Contains(t, res, "StructType")
}

// TestStringifyJSON_MarshalAvoided isolates the fix's effect from VM
// evaluation: given an already-built large string result, the unbounded path
// (pre-fix behavior) walks + amino-marshals it, churning many times the
// result size; the bounded path (the fix) aborts in the export walk and never
// reaches amino.MarshalJSON, so it allocates orders of magnitude less. The
// string is built once, before measurement, so neither delta includes it.
//
// Budgets are passed explicitly rather than via maxQueryExportBytes: this
// asserts stringifyJSONResults' own behavior, and the wiring to the production
// budget is covered by the two tests above.
func TestStringifyJSON_MarshalAvoided(t *testing.T) {
	const n = 2 << 20 // 2MB: enough for a decisive ratio, small enough to be cheap
	tvs := []gnolang.TypedValue{{
		T: gnolang.StringType,
		V: gnolang.StringValue(strings.Repeat("A", n)),
	}}
	// A plain string never triggers error extraction, so the nil machine is
	// unused on this path.
	var m *gnolang.Machine

	measure := func(fn func()) uint64 {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		fn()
		runtime.ReadMemStats(&after)
		return after.TotalAlloc - before.TotalAlloc
	}

	unbounded := measure(func() {
		s, err := stringifyJSONResults(m, tvs, nil, 0) // no bound: marshals
		require.NoError(t, err)
		require.Greater(t, len(s), n) // the full JSON was produced
	})

	bounded := measure(func() {
		_, err := stringifyJSONResults(m, tvs, nil, 100_000)
		require.ErrorIs(t, err, gnolang.ErrExportSizeExceeded) // aborts before marshal
	})

	// TotalAlloc is a process-global counter, so only the order-of-magnitude
	// ratio is asserted — an absolute ceiling here would flake on any
	// concurrent allocation.
	t.Logf("unbounded (pre-fix) alloc=%d bytes; bounded (fix) alloc=%d bytes", unbounded, bounded)
	assert.Less(t, bounded, unbounded/10,
		"bounded path should allocate far less; marshal was not avoided")
}

// TestQueryEvalJSON_BoundMethodNameGuarded exercises the BoundMethodValue.Method
// charge through the wired qeval_json endpoint: a slice of lazy interface-bound
// method values with a long method-name identifier must be refused, since the
// name is re-emitted per element by amino outside the eval meters.
func TestQueryEvalJSON_BoundMethodNameGuarded(t *testing.T) {
	const pkgPath = "gno.land/r/bmv"
	lowerExportBudget(t, 100_000)

	method := "M" + strings.Repeat("x", 5_000)
	env := setupSizeGuardRealm(t, "bmvaddr", pkgPath, `package bmv

type I interface{ `+method+`() int }

type Impl struct{ v int }

func (i Impl) `+method+`() int { return i.v }

func Many(n int) []func() int {
	var o I = Impl{v: 1}
	fns := make([]func() int, n)
	for k := range fns {
		fns[k] = o.`+method+`
	}
	return fns
}
`)

	// 400 elements each re-emit the ~5KB name (~2MB of JSON), but pre-fix only
	// the ~150B/node was charged (~60KB total), so this passed the 100KB budget.
	_, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Many(400)")
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})

	// A single method value of the same shape is well under budget and resolves.
	res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Many(1)")
	require.NoError(t, err)
	require.Contains(t, res, "results")
}

// TestQueryEvalJSON_AtErrorBounded is the regression test for the @error
// bypass: stringifyJSONResults bounds jres.Results via the export walk, but the
// realm's own Error() output feeds the @error field outside that budget. A
// realm whose Error() builds a large string must not force an unmetered
// multi-hundred-MB response through qeval_json's json.Marshal; the string is
// bounded like every other diagnostic string in the module (truncate).
func TestQueryEvalJSON_AtErrorBounded(t *testing.T) {
	const pkgPath = "gno.land/r/aterr"
	env := setupSizeGuardRealm(t, "aterraddr", pkgPath, `package aterr

import "strings"

type bigErr struct{ n int }

func (e bigErr) Error() string { return strings.Repeat("E", e.n) }

// Bad returns an error whose Error() text is far larger than any diagnostic
// bound; pre-fix its 300KB Error() string landed verbatim in @error.
func Bad() error { return bigErr{n: 300_000} }
`)

	res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Bad()")
	require.NoError(t, err)
	// The realm builds a 300KB Error() string, but @error is truncated, so the
	// whole response stays small.
	require.Less(t, len(res), 20_000,
		"unbounded @error must not dominate the response: got %d bytes", len(res))
	require.Contains(t, res, "@error")
}
