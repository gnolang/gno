package vm

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bigvalRealmSource returns functions that build large ephemeral values, used to
// reproduce the unmetered export-walk + JSON-marshal blowup on the JSON query
// endpoints and to verify the maxQueryExportBytes and maxQueryOutputBytes bounds.
const bigvalRealmSource = `package bigval

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

type Link struct {
	Next *Link
}

// Deep builds an n-deep ephemeral linked list. It is thin (one pointer per
// node), so it stays under the export byte budget even at tens of thousands of
// levels — but exporting and amino-marshaling it recurses n deep on the
// goroutine stack (a fatal, unrecoverable overflow) and the marshal is
// ~O(n^2). This is the vector the export depth guard bounds.
func Deep(n int) *Link {
	var head *Link
	for i := 0; i < n; i++ {
		head = &Link{Next: head}
	}
	return head
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

// lowerOutputBudget temporarily sets maxQueryOutputBytes, the response cap, under
// the same rules as lowerExportBudget. n <= 0 disables the writer caps, so a
// test can show that a refusal comes from the export estimate alone.
func lowerOutputBudget(t *testing.T, n int64) {
	t.Helper()
	prev := maxQueryOutputBytes
	maxQueryOutputBytes = n
	t.Cleanup(func() { maxQueryOutputBytes = prev })
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

// TestQueryEval_SizeGuard is the regression test for the memory exhaustion on
// both eval endpoints: a free query whose result would build a response past
// its limit must be rejected, not serviced. qeval_json is refused by the export
// estimate before amino runs, so the output cap is disabled there. qeval runs no
// export walk: TypedValue.String() streams into the output cap and stops at it.
func TestQueryEval_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/bigval"
	env := setupSizeGuardRealm(t, "dosaddr", pkgPath, bigvalRealmSource)

	// Both exprs are comfortably under the eval alloc/gas ceiling, so pre-fix
	// they were serviced and serialized unmetered.
	exprs := []string{"Big(200000)", "Structs(2000)"}
	t.Run("qeval_json", func(t *testing.T) {
		lowerExportBudget(t, 50_000)
		lowerOutputBudget(t, 0)
		for _, expr := range exprs {
			_, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, expr)
			require.Error(t, err, expr)
			assert.ErrorIs(t, err, ExportSizeExceededError{}, expr)
			// The log names the limit: a client sees which one refused.
			assert.Contains(t, fmt.Sprintf("%#v", err), "estimated response size exceeds 50000 bytes", expr)
		}
	})
	t.Run("qeval", func(t *testing.T) {
		lowerExportBudget(t, 0)
		lowerOutputBudget(t, 50_000)
		for _, expr := range exprs {
			_, err := env.vmk.QueryEval(env.ctx, pkgPath, expr)
			require.Error(t, err, expr)
			assert.ErrorIs(t, err, ExportSizeExceededError{}, expr)
			assert.Contains(t, fmt.Sprintf("%#v", err), "response size exceeds 50000 bytes", expr)
			assert.NotContains(t, fmt.Sprintf("%#v", err), "estimated", expr)
		}
	})

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
	const pkgPath = "gno.land/r/bigobj"
	lowerExportBudget(t, 20_000)

	// 100KB persisted byte array: charged at the base64 rate (~133KB), well
	// over the lowered budget once the array object itself is expanded.
	env := setupSizeGuardRealm(t, "bigobjaddr", pkgPath, `package bigobj

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

// TestQueryEval_DepthGuard covers a thin, deeply nested value that stays under
// the export byte budget. The export walk and amino marshal recurse on the
// goroutine stack, where an overflow is fatal and recover() cannot catch it,
// and the marshal is superlinear in depth. qeval_json must reject it with
// ExportDepthExceededError and keep serving queries.
//
// qeval runs no export walk: TypedValue.String() stops descending at
// nestedLimit, so the same value renders to a short, truncated response.
//
// Uses the real (unlowered) budgets: the depth guard is a fixed cap,
// independent of the byte budgets, and the point is that this shape slips under
// them.
func TestQueryEval_DepthGuard(t *testing.T) {
	const pkgPath = "gno.land/r/bigval" // must match bigvalRealmSource's `package bigval`
	env := setupSizeGuardRealm(t, "dosdepthaddr", pkgPath, bigvalRealmSource)

	res, err := env.vmk.QueryEval(env.ctx, pkgPath, "Deep(5000)")
	require.NoError(t, err)
	require.Less(t, len(res), 4096)
	require.Contains(t, res, "...")

	// Well past maxExportDepth but only a few hundred KB of estimated export
	// size, so the depth guard fires, not the size guard.
	_, err = env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Deep(5000)")
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportDepthExceededError{})

	// The node still serves queries after the guard fires.
	res, err = env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Deep(5)")
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

// TestQueryPkgJSON_SizeGuard verifies the qpkg_json path is bounded too: a
// package variable holding a large value must not force an unmetered export +
// marshal.
func TestQueryPkgJSON_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/bigpkg"
	lowerExportBudget(t, 20_000)

	// 100KB var: over the lowered export budget, so the qpkg_json export walk
	// must reject it.
	env := setupSizeGuardRealm(t, "bigpkgaddr", pkgPath, `package bigpkg

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
// JSON would blow past the output cap must stop at it, while an ordinary leaf
// type still resolves. QueryType is called directly, so the handler's cap on
// the final bytes is not what refuses it.
func TestQueryType_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/amp"
	lowerOutputBudget(t, 50_000)
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
// Budgets are passed explicitly rather than via the package vars: this
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
		s, err := stringifyJSONResults(m, tvs, nil, 0, 0) // no bound: marshals
		require.NoError(t, err)
		require.Greater(t, len(s), n) // the full JSON was produced
	})

	bounded := measure(func() {
		_, err := stringifyJSONResults(m, tvs, nil, 100_000, 0)
		require.ErrorIs(t, err, gnolang.ErrExportSizeExceeded) // aborts before marshal
	})

	// TotalAlloc is a process-global counter, so only the order-of-magnitude
	// ratio is asserted — an absolute ceiling here would flake on any
	// concurrent allocation.
	t.Logf("unbounded (pre-fix) alloc=%d bytes; bounded (fix) alloc=%d bytes", unbounded, bounded)
	assert.Less(t, bounded, unbounded/10,
		"bounded path should allocate far less; marshal was not avoided")
}

// TestQueryRender_RefusedBeforeCopy checks that an oversized Render result is
// refused before qrender copies it into the response: the handler call costs
// about what the keeper call alone does, not that plus another copy.
func TestQueryRender_RefusedBeforeCopy(t *testing.T) {
	const n = 32 << 20
	const pkgPath = "gno.land/r/bigrender"
	env := setupSizeGuardRealm(t, "bigrenderaddr", pkgPath, fmt.Sprintf(`package bigrender

import "strings"

func Render(string) string { return strings.Repeat("x", %d) }
`, n))

	measure := func(fn func()) uint64 {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		fn()
		runtime.ReadMemStats(&after)
		return after.TotalAlloc - before.TotalAlloc
	}

	keeper := measure(func() {
		s, err := env.vmk.QueryEvalString(env.ctx, pkgPath, `Render("")`)
		require.NoError(t, err)
		require.Len(t, s, n)
	})
	handler := measure(func() {
		res := NewHandler(env.vmk).Query(env.ctx, abci.RequestQuery{Path: "vm/qrender", Data: []byte(pkgPath + ":")})
		require.Error(t, res.Error)
		assert.ErrorIs(t, res.Error, ExportSizeExceededError{})
	})

	// TotalAlloc is process-global, so only a margin well below n is asserted.
	t.Logf("keeper alloc=%d bytes; handler alloc=%d bytes", keeper, handler)
	assert.Less(t, handler, keeper+n/2, "qrender copied the result before refusing it")
}

// TestQueryEvalJSON_BoundMethodNameGuarded exercises the BoundMethodValue.Method
// charge through the wired qeval_json endpoint: a slice of lazy interface-bound
// method values with a long method-name identifier must be refused, since the
// name is re-emitted per element by amino outside the eval meters.
func TestQueryEvalJSON_BoundMethodNameGuarded(t *testing.T) {
	const pkgPath = "gno.land/r/bmv"
	lowerExportBudget(t, 100_000)
	lowerOutputBudget(t, 0) // the estimate alone must charge the name

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

// gnomodBoundsCases are the gnomod.toml bodies that an unbounded TOML decoder turns
// into a validator-wide stall. All are caller supplied: they ride in the
// MemPackage of a MsgAddPackage or MsgRun, and both the type-check pass and the
// keeper's own checks decode them.
//
// Two layers reject these. gnomod.maxFileSize caps length (the first case);
// tm2/pkg/toml caps nesting depth and key-path depth against the parser's own
// recursion counter and its own parsed key paths (every other case, all of which
// are under the size bound so they exercise the decoder). Unit-level equivalents
// live in gnovm/pkg/gnomod and tm2/pkg/toml; what this table adds is that the
// bounds are actually reached from both consensus entry points.
var gnomodBoundsCases = []struct {
	name    string
	body    func(pkgPath string) string
	wantErr string
}{
	{
		// parseGroup rescans seenTableKeys per table, so decoding
		// is O(n²) in the table count — a 1MB body took ~9s per decode, twice per
		// message, against 10 gas/byte of tx-size gas and nothing else.
		name: "table count",
		body: func(pkgPath string) string {
			var b strings.Builder
			fmt.Fprintf(&b, "module = %q\ngno = \"0.9\"\n", pkgPath)
			for i := 0; b.Len() < 700_000; i++ {
				fmt.Fprintf(&b, "[t%06d]\nk=1\n", i)
			}
			return b.String()
		},
		wantErr: "exceeds limit 4096",
	},
	{
		// parseRvalue/parseArray recurse per nesting level with no
		// depth limit, so under 400KB of '[' exhausts Go's 1GB goroutine stack. A stack
		// overflow is fatal — no recover() catches it — so every node processing
		// the message dies and the chain halts. Kept under maxFileSize here so it
		// is the delimiter bound, not the size bound, that rejects it.
		name: "nesting depth",
		body: func(pkgPath string) string {
			return fmt.Sprintf("module = %q\ngno = \"0.9\"\na = %s\n",
				pkgPath, strings.Repeat("[", 600)+strings.Repeat("]", 600))
		},
		wantErr: "nesting depth exceeds limit 256",
	},
	{
		// The same nesting, with the closers moved into comments where the decoder
		// ignores them. A guard counting brackets in the raw bytes reads +1 per line
		// here while the decoder still recurses +10 — one of the reasons this bound
		// belongs in the parser, counting actual recursion rather than approximating
		// it from text.
		name: "nesting behind comments",
		body: func(pkgPath string) string {
			return fmt.Sprintf("module = %q\ngno = \"0.9\"\na = %s",
				pkgPath, strings.Repeat(strings.Repeat("[", 10)+"#"+strings.Repeat("]", 9)+"\n", 60))
		},
		wantErr: "nesting depth exceeds limit 256",
	},
	{
		// The axis no delimiter reveals: parseAssign re-walks the whole current
		// table key path for every assignment, so a deep dotted table header with
		// many assignments under it is O(n·d) — quadratic in the body — while
		// spending one single '['. Under maxFileSize (so the delimiter and size
		// bounds both pass) this decoded in 2.8ms, 11x the worst those two admit.
		name: "dotted key-path depth",
		body: func(pkgPath string) string {
			var b strings.Builder
			fmt.Fprintf(&b, "module = %q\ngno = \"0.9\"\n", pkgPath)
			fmt.Fprintf(&b, "[%sz]\n", strings.Repeat("a.", 1015))
			for i := 0; ; i++ {
				line := fmt.Sprintf("%c%c%c = 1\n", 'a'+i/676%26, 'a'+i/26%26, 'a'+i%26)
				if b.Len()+len(line) > 4096 {
					break
				}
				b.WriteString(line)
			}
			return b.String()
		},
		wantErr: "key path depth",
	},
	{
		// The same O(n·d) axis bought with no dot at all: parseKey appends a key
		// group on every quoted run and does not require a '.' between segments
		// (keysparsing.go), so ["a""b""c"] is the path a.b.c and a quoted header
		// buys one level of depth per 2 bytes. A dot count cannot see this at all
		// — measured, 4KB of it reached ~2000 levels at 2 dots in the whole body,
		// decoding in 2.8ms against the 609µs the dotted shape was believed to
		// cap, on the same ~4x-per-doubling curve (988ms at 64KB) and ~1391 gas of
		// work per byte across the two decodes against the 1250 charged. A dot count
		// admitted this outright; bounding parseKey's own parsed path is what
		// rejects it, since that path is a.b.c however it was spelled.
		name: "quoted key-path depth, zero dots",
		body: func(pkgPath string) string {
			var b strings.Builder
			fmt.Fprintf(&b, "module = %q\ngno = \"0.9\"\n", pkgPath)
			fmt.Fprintf(&b, "[%s]\n", strings.Repeat(`""`, 1000))
			for i := 0; ; i++ {
				line := fmt.Sprintf("%c%c%c = 1\n", 'a'+i/676%26, 'a'+i/26%26, 'a'+i%26)
				if b.Len()+len(line) > 4096 {
					break
				}
				b.WriteString(line)
			}
			return b.String()
		},
		wantErr: "key path depth",
	},
	{
		// The same O(n·d) axis, reached without ever putting many dots on one
		// line: go-toml's lexInsideTableKey copies bytes through to parseKey
		// until ']', and parseKey's quoted-key branch consumes runes to the
		// closing quote without checking for a newline, so a table header can
		// span lines inside a quoted segment. That buys one level of depth per 4
		// bytes at a single dot per line, which a per-line byte count cannot see —
		// and requiring balanced quotes per line does not save it either, since a
		// '"' inside a single-quoted segment or a trailing comment pads the parity
		// without delimiting anything. Measured unbounded: 4093 bytes reached 1333 levels
		// and cost 11ms and 40MB per message against the ~5.2M gas charged over
		// them — a full block of them was ~5.3s of CPU and ~19GB allocated,
		// since the gas ceiling does not bind before the 2MB block-data one.
		name: "key path spanning lines",
		body: func(pkgPath string) string {
			var b strings.Builder
			fmt.Fprintf(&b, "module = %q\ngno = \"0.9\"\n", pkgPath)
			b.WriteString("[\"\"")
			for range 1000 {
				b.WriteString(".\"\n\"")
			}
			b.WriteString("]\n")
			return b.String()
		},
		wantErr: "key path depth",
	},
}

// TestAddPackage_GnoModBounds covers every oversized gnomod.toml body through
// MsgAddPackage: gnomod's bounds must reject them at the decode, well before
// any of the costs is paid.
func TestAddPackage_GnoModBounds(t *testing.T) {
	const pkgPath = "gno.land/r/gnomodbig"
	for _, tc := range gnomodBoundsCases {
		t.Run(tc.name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

			addr := crypto.AddressFromPreimage([]byte("gnomodbigaddr"))
			env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
			env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(100_000_000)))

			files := []*std.MemFile{
				{Name: "a.gno", Body: "package gnomodbig\n\nfunc Echo() string { return \"ok\" }\n"},
				{Name: "gnomod.toml", Body: tc.body(pkgPath)},
			}
			err := env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files))
			require.Error(t, err)
			// ErrTypeCheck keeps only the sentinel on the hashed path; the
			// message rides the msg trace (see TestErrTypeCheckCoarseHashedError).
			assert.Contains(t, fmt.Sprintf("%#v", err), tc.wantErr)
		})
	}
}

// TestRun_GnoModBounds is the MsgRun half: Run type-checks the caller's package —
// decoding their gnomod.toml — before overwriting it with a generated one, so it
// reaches the same decoder as MsgAddPackage.
func TestRun_GnoModBounds(t *testing.T) {
	for _, tc := range gnomodBoundsCases {
		t.Run(tc.name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

			addr := crypto.AddressFromPreimage([]byte("gnomodbigaddr"))
			env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
			env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(100_000_000)))

			pkgPath := "gno.land/e/" + addr.String() + "/run"
			files := []*std.MemFile{
				{Name: "gnomod.toml", Body: tc.body(pkgPath)},
				{Name: "script.gno", Body: "package main\n\nfunc main() {}\n"},
			}
			_, err := env.vmk.Run(ctx, NewMsgRun(addr, std.MustParseCoins(""), files))
			require.Error(t, err)
			assert.Contains(t, fmt.Sprintf("%#v", err), tc.wantErr)
		})
	}
}

// TestChargePreprocessGas_GnoModCharged pins that the mod file is metered like
// .gno source. Its body is decoded twice per message and was previously skipped
// here entirely, leaving the ante handler's 10 gas/byte as the only charge
// against an unbounded decoder — a ~125x discount on caller-controlled bytes.
func TestChargePreprocessGas_GnoModCharged(t *testing.T) {
	env := setupTestEnv()
	params := env.vmk.GetParams(env.ctx)

	charge := func(files ...*std.MemFile) int64 {
		ctx, _ := env.ctx.WithGasMeter(store.NewGasMeter(DefaultQueryMaxGas)).CacheContext()
		chargePreprocessGas(ctx, params, &std.MemPackage{Files: files}, "test")
		return ctx.GasMeter().GasConsumed()
	}

	src := &std.MemFile{Name: "a.gno", Body: "package a\n"}
	mod := &std.MemFile{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/a")}
	dotMod := &std.MemFile{Name: "gno.mod", Body: "module gno.land/r/a\n"}
	readme := &std.MemFile{Name: "README.md", Body: strings.Repeat("x", 4096)}

	perByte := params.PreprocessGasPerByte
	assert.Equal(t, perByte*int64(len(src.Body)), charge(src))
	assert.Equal(t, perByte*int64(len(src.Body)+len(mod.Body)), charge(src, mod))
	assert.Equal(t, perByte*int64(len(src.Body)+len(dotMod.Body)), charge(src, dotMod))
	// Files no decoder reads stay free; the tx-size gas already covers them.
	assert.Equal(t, perByte*int64(len(src.Body)), charge(readme, src))
}

// TestQueryEval_SharedSubtreeBounded covers values that are small in memory but
// large when rendered. The export walk counts a shared object once, while
// String() renders it at every reference. qeval must be bounded by the bytes
// it renders.
func TestQueryEval_SharedSubtreeBounded(t *testing.T) {
	const pkgPath = "gno.land/r/dagval"
	lowerExportBudget(t, 0)
	lowerOutputBudget(t, 50_000)
	env := setupSizeGuardRealm(t, "dagaddr", pkgPath, `package dagval

type Dag struct{ Kids []*Dag }

func MkDag(depth, fanout int) *Dag {
	d := &Dag{}
	for i := 0; i < depth; i++ {
		p := &Dag{Kids: make([]*Dag, fanout)}
		for j := range p.Kids {
			p.Kids[j] = d
		}
		d = p
	}
	return d
}

// Wide is two objects: one slice whose n elements all share one []int of
// length n. It renders as n² ints, while the export estimate counts 2n nodes.
func Wide(n int) []any {
	child := make([]int, n)
	arr := make([]any, n)
	for i := range arr {
		arr[i] = child
	}
	return arr
}
`)

	// Two objects that render to ~720KB, bounded by neither depth nor object
	// count. The export estimate let Wide(8000) through as a 512MB response.
	_, err := env.vmk.QueryEval(env.ctx, pkgPath, "Wide(300)")
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})

	// ~60 objects in memory, ~27K leaves (over 1MB) rendered.
	_, err = env.vmk.QueryEval(env.ctx, pkgPath, "MkDag(3,30)")
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})

	// The JSON form exports each shared object once, so it stays small.
	_, err = env.vmk.QueryEvalJSON(env.ctx, pkgPath, "MkDag(3,30)")
	require.NoError(t, err)

	res, err := env.vmk.QueryEval(env.ctx, pkgPath, "MkDag(3,2)")
	require.NoError(t, err)
	assert.Contains(t, res, "Dag")
}

// TestQueryEval_TypeStringBounded covers a type that is small in memory but
// large when rendered: aliased anonymous structs of 12 fields, 6 deep, render
// to 12^6 leaves. qeval must stop the type string at the output cap too, not
// build it whole first (917MB allocated before the fix).
func TestQueryEval_TypeStringBounded(t *testing.T) {
	const pkgPath = "gno.land/r/typestr"
	env := setupSizeGuardRealm(t, "typestraddr", pkgPath, "package typestr\n\nvar X int\n")

	names := make([]string, 12)
	for i := range names {
		names[i] = fmt.Sprintf("f%d", i)
	}
	var types strings.Builder
	types.WriteString("type A0 = int; ")
	for i := 1; i <= 6; i++ {
		fmt.Fprintf(&types, "type A%d = struct{ %s A%d }; ", i, strings.Join(names, ", "), i-1)
	}

	// The closure, bound method and type value forms are covered in gnolang.
	expr := "func() any { " + types.String() + "var v *A6; return v }()"
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := env.vmk.QueryEval(env.ctx, pkgPath, expr)
	runtime.ReadMemStats(&after)
	require.Error(t, err)
	assert.ErrorIs(t, err, ExportSizeExceededError{})
	// The render stops at the 1 MiB cap. TotalAlloc is process-global, so the
	// bound is loose.
	assert.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(200<<20))
}

// TestQueryOutputCap_Streaming covers the writer caps of the amino JSON
// endpoints. The export estimate is disabled and each keeper method is called
// directly, past the handler, so the refusal can only come from the
// cappedBuffer amino writes into. Half the full answer still holds every
// envelope, so the cap is below the inner JSON it bounds.
func TestQueryOutputCap_Streaming(t *testing.T) {
	const pkgPath = "gno.land/r/ptrs"
	env := setupSizeGuardRealm(t, "ptrsaddr", pkgPath, `package ptrs

type T struct{ N int }

type Rec struct {
	ID   int
	Name string
}

var List = make([]*T, 200)

var All = [50]Rec{}

func init() {
	for i := range List {
		List[i] = &T{N: i}
	}
}

func GetList() []*T { return List }
`)
	res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetList()")
	require.NoError(t, err)
	oid := extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "no ObjectID in %s", res)

	for _, tc := range []struct {
		name  string
		query func() (string, error)
	}{
		{"qeval_json", func() (string, error) { return env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetList()") }},
		{"qobject_json", func() (string, error) { return env.vmk.QueryObjectJSON(env.ctx, oid) }},
		{"qpkg_json", func() (string, error) { return env.vmk.QueryPkg(env.ctx, pkgPath) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			full, err := tc.query()
			require.NoError(t, err)

			lowerExportBudget(t, 0)
			lowerOutputBudget(t, int64(len(full))/2)
			_, err = tc.query()
			assert.ErrorIs(t, err, ExportSizeExceededError{})
			assert.Contains(t, fmt.Sprintf("%#v", err), fmt.Sprintf("response size exceeds %d bytes", len(full)/2))
		})
	}
}

// TestQueryFuncs_SizeGuard covers the cumulative budget of qfuncs' type text.
// Each parameter of Wide renders the same 4^4-leaf type, so a budget of twice
// its length fits both, one byte less stops the second in its writer, and a
// budget the first uses up exactly stops the second before it starts.
func TestQueryFuncs_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/widefunc"
	env := setupSizeGuardRealm(t, "widefuncaddr", pkgPath,
		"package widefunc\n\n"+typeChain(4, 4)+"\nfunc Wide(v, w T4) {}\n")
	fsigs, err := env.vmk.QueryFuncs(env.ctx, pkgPath)
	require.NoError(t, err)
	n := int64(len(fsigs[0].Params[0].Type))
	lowerExportBudget(t, 0)

	lowerOutputBudget(t, 2*n)
	_, err = env.vmk.QueryFuncs(env.ctx, pkgPath)
	require.NoError(t, err)

	for _, budget := range []int64{2*n - 1, n} {
		lowerOutputBudget(t, budget)
		_, err = env.vmk.QueryFuncs(env.ctx, pkgPath)
		assert.ErrorIs(t, err, ExportSizeExceededError{}, "budget %d", budget)
	}
}

// TestQueryHandler_StreamingJSON covers the writer caps qfuncs and qdoc marshal
// into. Their handler functions are called directly, past the final cap in
// vmHandler.Query, with the output cap at half the full answer.
func TestQueryHandler_StreamingJSON(t *testing.T) {
	const pkgPath = "gno.land/r/docs"
	env := setupSizeGuardRealm(t, "docsaddr", pkgPath, `package docs

// Hello greets.
func Hello(name string, times int) (string, error) { return name, nil }

// Bye says goodbye.
func Bye(name string) string { return name }
`)
	for _, tc := range []struct {
		name  string
		query func(sdk.Context, abci.RequestQuery) abci.ResponseQuery
	}{
		{"qfuncs", env.vmh.queryFuncs},
		{"qdoc", env.vmh.queryDoc},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := abci.RequestQuery{Data: []byte(pkgPath)}
			full := tc.query(env.ctx, req)
			require.True(t, full.IsOK(), "%v", full.Error)

			lowerOutputBudget(t, int64(len(full.Data))/2)
			res := tc.query(env.ctx, req)
			require.False(t, res.IsOK())
			assert.IsType(t, ExportSizeExceededError{}, res.Error)
			assert.Empty(t, res.Data)
		})
	}
}

// TestCappedBuffer_AminoJSON checks that amino.MarshalJSONTo into a
// cappedBuffer fails at every cap below the full size, rather than returning
// success with truncated JSON.
func TestCappedBuffer_AminoJSON(t *testing.T) {
	v := []string{"a", "bb", "ccc"}
	full, err := amino.MarshalJSON(v)
	require.NoError(t, err)
	for n := int64(1); n < int64(len(full)); n++ {
		cb := cappedBuffer{max: n}
		err := amino.MarshalJSONTo(&cb, v)
		require.ErrorIs(t, err, gnolang.ErrExportSizeExceeded, "max=%d", n)
	}
	cb := cappedBuffer{max: int64(len(full))}
	require.NoError(t, amino.MarshalJSONTo(&cb, v))
	require.Equal(t, full, cb.buf.Bytes())

	// Once refused, a shorter write is refused too.
	cb = cappedBuffer{max: 3}
	_, err = cb.Write([]byte("abcd"))
	require.ErrorIs(t, err, gnolang.ErrExportSizeExceeded)
	_, err = cb.Write([]byte("a"))
	require.ErrorIs(t, err, gnolang.ErrExportSizeExceeded)
	require.Zero(t, cb.buf.Len())
}

// TestQueryEvalJSON_AtErrorBudgetFails checks that a budget exhausted while
// extracting @error fails the query instead of swallowing the panic and
// returning a false success with no @error. The call returns an error value
// cheaply, then Error() allocates past QueryMaxAlloc, so the allocator panics
// during extraction.
func TestQueryEvalJSON_AtErrorBudgetFails(t *testing.T) {
	const pkgPath = "gno.land/r/aterrbudget"
	env := setupSizeGuardRealm(t, "aterrbudgetaddr", pkgPath, `package aterrbudget

type allocErr struct{}

func (allocErr) Error() string {
	// Allocate well past the lowered query budget only when Error() runs.
	sink := make([]byte, 60_000_000)
	return string(sink[:1])
}

// Bad returns cheaply. Only its Error() exhausts the allocation budget.
func Bad() error { return allocErr{} }
`)

	env.vmk.QueryMaxAlloc = 20_000_000

	res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Bad()")
	require.Error(t, err,
		"@error extraction that exhausts the allocation budget must fail the query")
	require.Contains(t, err.Error(), "allocation limit exceeded")
	require.Empty(t, res)
}

// TestQueryDoc_SizeGuard checks vm/qdoc against a field whose names all share
// one long type: the JSON repeats the type once per name, so
// it is refused past maxQueryOutputBytes, and a small doc is unchanged.
func TestQueryDoc_SizeGuard(t *testing.T) {
	const pkgPath = "gno.land/r/widedoc"
	names := make([]string, 120)
	for i := range names {
		names[i] = fmt.Sprintf("F%d", i)
	}
	// One AST node but many bytes, so the declaration stays cheap to deploy.
	typ := "L" + strings.Repeat("o", 1000) + "ng"
	env := setupSizeGuardRealm(t, "docaddr", pkgPath, "package widedoc\n\ntype "+typ+" int\n\ntype T struct {\n\t"+
		strings.Join(names, ", ")+" "+typ+"\n}\n")
	query := func() abci.ResponseQuery {
		return env.vmh.Query(env.ctx, abci.RequestQuery{Path: "vm/qdoc", Data: []byte(pkgPath)})
	}

	// Unchanged output under the default budget.
	res := query()
	require.True(t, res.IsOK(), "%v", res.Error)
	jsonDoc, err := env.vmk.QueryDoc(env.ctx, pkgPath)
	require.NoError(t, err)
	assert.Equal(t, jsonDoc.JSON(), string(res.Data))

	// 120 names × a ~1000 byte type is over 100KB of JSON.
	lowerOutputBudget(t, 50_000)
	res = query()
	require.False(t, res.IsOK())
	assert.IsType(t, ExportSizeExceededError{}, res.Error)
}

// typeChain declares T0 and T1 to T<levels>, each a struct of width fields of the
// level below, one per line.
func typeChain(levels, width int) string {
	decls := []string{"type T0 struct{ V int }"}
	for l := 1; l <= levels; l++ {
		var fields strings.Builder
		for f := range width {
			fields.WriteString(fmt.Sprintf("F%d T%d; ", f, l-1))
		}
		decls = append(decls, fmt.Sprintf("type T%d = struct{ %s}", l, fields.String()))
	}
	return strings.Join(decls, "\n")
}
