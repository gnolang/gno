package gnolang

import (
	"math/big"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// TestExportValuesLimit_BoundMethodNameCharged is the regression test for the
// BoundMethodValue.Method / MethodPkg bypass: a slice of lazy interface-bound
// method values (Func == nil) re-emits the caller-controlled method-name
// identifier once per element, so the walk must charge it or a long method name
// slips past the budget — the same field-name/tag bypass class as
// TestExportValuesLimit_FieldNamesAndTagsCharged, on a different node.
func TestExportValuesLimit_BoundMethodNameCharged(t *testing.T) {
	const (
		nameLen = 4_000
		n       = 200
	)
	method := "M" + strings.Repeat("x", nameLen-1)

	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata

type I interface{ `+method+`() int }

type Impl struct{ v int }

func (i Impl) `+method+`() int { return i.v }

func F(n int) []func() int {
	var o I = Impl{v: 1}
	fns := make([]func() int, n)
	for k := range fns {
		fns[k] = o.`+method+`
	}
	return fns
}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Call(Sel(Nx("testdata"), "F"), X(n)))

	// The emitted method-name content (n × nameLen = 800KB) is now charged, so
	// a 100KB budget rejects this. Pre-fix the names were charged 0 and this
	// sailed through.
	_, err := ExportValues(tps, 100_000)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	// With a budget above the name content it is accepted, and the marshaled
	// output stays within a small factor of the budget — the property the
	// charge exists to provide.
	const budget = 4_000_000
	exported, err := ExportValues(tps, budget)
	require.NoError(t, err)
	bz, err := amino.MarshalJSON(exported)
	require.NoError(t, err)
	require.Less(t, len(bz), budget*2,
		"marshaled output (%d B) should stay within 2x the %d B budget", len(bz), budget)
}

func TestExportValuesLimit_BoundMethodReceiverDepthGuard(t *testing.T) {
	build := func(depth int) []TypedValue {
		receiver := TypedValue{T: IntType}
		for range depth {
			receiver = TypedValue{
				T: &FuncType{},
				V: &BoundMethodValue{Receiver: receiver, Method: "M"},
			}
		}
		return []TypedValue{receiver}
	}

	_, err := ExportValues(build(10), 100_000_000)
	require.NoError(t, err)
	_, err = ExportValues(build(maxExportDepth*2), 100_000_000)
	require.ErrorIs(t, err, ErrExportDepthExceeded)
}

// TestExportValuesLimit_String verifies that a large ephemeral string is
// rejected by the size bound before the export walk completes, while a small
// one passes.
func TestExportValuesLimit_String(t *testing.T) {
	tv := TypedValue{T: StringType, V: StringValue{Str: strings.Repeat("A", 1_000_000)}}

	// Under budget: passes.
	got, err := ExportValues([]TypedValue{tv}, 2_000_000)
	require.NoError(t, err)
	require.Len(t, got, 1)

	// Over budget: rejected before building the full copy.
	_, err = ExportValues([]TypedValue{tv}, 100_000)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	// maxBytes <= 0 disables the bound, for trusted callers.
	got, err = ExportValues([]TypedValue{tv}, 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
}

// TestExportValuesLimit_ManyNodes verifies the per-node charge catches a
// tree that is huge by node count rather than by a single large scalar.
func TestExportValuesLimit_ManyNodes(t *testing.T) {
	// A list-backed array of many small int elements. An int element costs ~96
	// bytes of budget, so 20_000 sits ~20x over the 100_000 budget below.
	const n = 20_000
	list := make([]TypedValue, n)
	for i := range list {
		list[i] = TypedValue{T: IntType, N: [8]byte{}}
	}
	arr := &ArrayValue{List: list}
	tv := TypedValue{T: &SliceType{Elt: IntType}, V: arr}

	_, err := ExportValues([]TypedValue{tv}, 100_000)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	// Generous budget: passes.
	_, err = ExportValues([]TypedValue{tv}, int64(n)*exportNodeEst*4)
	require.NoError(t, err)
}

// TestExportValuesLimit_FieldNamesAndTagsCharged is the regression test for the
// field-name/struct-tag bypass described in exportNodeEst's doc comment.
//
// The shape below is deliberately tiny (400KB of tag content). Pre-fix it was
// charged ~16KB and sailed through a 100KB budget; it must now be rejected, and
// whatever is accepted must marshal to within a small factor of the budget.
func TestExportValuesLimit_FieldNamesAndTagsCharged(t *testing.T) {
	const (
		tagLen = 4_000
		n      = 100
	)

	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata

type Elem = struct {
	A int "`+strings.Repeat("T", tagLen)+`"
}

func F(n int) []Elem { return make([]Elem, n) }
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Call(Sel(Nx("testdata"), "F"), X(n)))

	// The emitted tag content (n × tagLen = 400KB) is now charged, so a 100KB
	// budget rejects this.
	_, err := ExportValues(tps, 100_000)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	// With a budget above the tag content it is accepted, and the marshaled
	// output stays within a small factor of the budget — the property the bound
	// exists to provide.
	const budget = 2_000_000
	exported, err := ExportValues(tps, budget)
	require.NoError(t, err)
	bz, err := amino.MarshalJSON(exported)
	require.NoError(t, err)
	require.Less(t, len(bz), budget*2,
		"marshaled output (%d B) should stay within 2x the %d B budget", len(bz), budget)
}

// TestExportValuesLimit_ByteArrayData pins the []byte charge: Data is emitted
// as base64, so it is charged at 4/3 its raw length, before cp() duplicates it.
// A []byte return value is one of the cheapest large results a realm can build,
// so this is the guard's charge, not a defensive one.
func TestExportValuesLimit_ByteArrayData(t *testing.T) {
	const dataLen = 300_000
	charged := int64(dataLen) * 4 / 3 // 400_000

	arr := &ArrayValue{Data: make([]byte, dataLen)}
	tv := TypedValue{
		T: &SliceType{Elt: Uint8Type},
		V: &SliceValue{Base: arr, Length: dataLen, Maxcap: dataLen},
	}

	// A budget above the raw length but below the base64 charge is rejected:
	// the estimate follows the emitted size, not the in-memory size.
	_, err := ExportValues([]TypedValue{tv}, charged-10_000)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	_, err = ExportValues([]TypedValue{tv}, charged+10_000)
	require.NoError(t, err)
}

// TestExportObjectLimit covers the second entry point, used by qobject_json and
// qobject_binary: ExportObject must be bounded exactly like ExportValues.
func TestExportObjectLimit(t *testing.T) {
	arr := &ArrayValue{Data: make([]byte, 1_000_000)}

	_, err := ExportObject(arr, 100_000)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	exported, err := ExportObject(arr, 10_000_000)
	require.NoError(t, err)
	require.NotNil(t, exported)

	// maxBytes <= 0 disables the bound here too.
	exported, err = ExportObject(arr, 0)
	require.NoError(t, err)
	require.NotNil(t, exported)
}

// TestExportValuesLimit_Bigint pins the BigintValue charge (BitLen/3) as an
// over-estimate of the decimal text amino emits: a budget large enough for
// every emitted digit must still be rejected.
func TestExportValuesLimit_Bigint(t *testing.T) {
	bi := new(big.Int).Lsh(big.NewInt(1), 300_000)
	emitted := int64(len(bi.String())) // ~90_361 digits
	charged := int64(bi.BitLen())/3 + 1

	require.Greater(t, charged, emitted, "the charge must over-estimate the emitted text")

	tv := TypedValue{T: UntypedBigintType, V: BigintValue{V: bi}}
	_, err := ExportValues([]TypedValue{tv}, emitted)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	_, err = ExportValues([]TypedValue{tv}, charged+1_000)
	require.NoError(t, err)
}

// TestExportValuesLimit_Bigdec pins the BigdecValue charge. Pre-fix the rat
// form was charged nothing (only exportNodeEst), so a rat whose RatString runs
// to hundreds of KB slipped through any budget; the charge must now over-
// estimate that emitted text, exactly as the BigintValue charge does.
func TestExportValuesLimit_Bigdec(t *testing.T) {
	num := new(big.Int).Lsh(big.NewInt(1), 300_000) // 2^300000
	// 2^300000 and 2^300000-1 are consecutive integers, hence coprime, so the
	// rat is already in lowest terms and keeps both large components.
	den := new(big.Int).Sub(num, big.NewInt(1))
	r := new(big.Rat).SetFrac(num, den)

	emitted := int64(len(r.RatString())) // "num/den" decimal, ~180K bytes
	charged := int64(r.Num().BitLen()+r.Denom().BitLen())/3 + 1

	require.Greater(t, charged, emitted, "the charge must over-estimate the emitted text")

	tv := TypedValue{T: UntypedBigdecType, V: BigdecValue{V: r}}
	_, err := ExportValues([]TypedValue{tv}, emitted)
	require.ErrorIs(t, err, ErrExportSizeExceeded)

	_, err = ExportValues([]TypedValue{tv}, charged+1_000)
	require.NoError(t, err)
}

// buildDeepList evaluates a package that builds a `depth`-deep ephemeral linked
// list (a struct with one self-pointer field per level) and returns the result
// TypedValues. This is the thin, deeply nested shape that stays cheap per node,
// so tens of thousands of levels fit under the byte budget — the shape the depth
// guard exists to bound.
func buildDeepList(t *testing.T, depth int) []TypedValue {
	t.Helper()
	m := NewMachine("testdata", nil)
	t.Cleanup(m.Release)
	nn := m.MustParseFile("deep.gno", `package testdata
type Node struct {
	Next *Node
}
func Build(depth int) *Node {
	var head *Node
	for i := 0; i < depth; i++ {
		head = &Node{Next: head}
	}
	return head
}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))
	return m.Eval(Call(Sel(Nx("testdata"), "Build"), X(depth)))
}

// TestExportValuesLimit_DepthGuard is the regression test for the deep-recursion
// Note: a thin, deeply nested value tree stays under the byte budget (its per-node
// charge is tiny), so the size guard never fires — but exporting and then
// amino-marshaling it recurses on the goroutine stack (fatal, unrecoverable
// overflow) and the marshal is ~O(depth^2). The depth guard must reject it with
// ErrExportDepthExceeded, and must do so during the walk (before amino runs).
func TestExportValuesLimit_DepthGuard(t *testing.T) {
	// Deep enough to clear maxExportDepth by a wide margin, yet with a byte
	// budget so generous the size guard cannot be what fires: this isolates the
	// depth guard. The per-node charge (~tens of bytes) keeps the whole tree far
	// under 100MB, so ErrExportSizeExceeded is not reachable here.
	deep := buildDeepList(t, maxExportDepth*3)
	_, err := ExportValues(deep, 100_000_000)
	require.ErrorIs(t, err, ErrExportDepthExceeded)
}

// TestExportValuesLimit_DepthDisabled confirms the depth bound, like the size
// bound, is only enforced under a limiter: a trusted caller (maxBytes <= 0)
// exports a tree nested past maxExportDepth without error. Kept just above the
// cap (not at the attack scale) so the unbounded walk itself is cheap.
func TestExportValuesLimit_DepthDisabled(t *testing.T) {
	deep := buildDeepList(t, maxExportDepth+500)
	_, err := ExportValues(deep, 0)
	require.NoError(t, err)
}

// TestExportObjectLimit_DepthGuard covers the other bounded entry point, the one
// vm/qobject_json and vm/qobject_binary share. It is tested here rather than
// through the keeper because those endpoints resolve a *persisted* ObjectID, and
// persisted children collapse to RefValue one level in — so a deep graph is only
// reachable through ExportObject directly, with an ephemeral object.
func TestExportObjectLimit_DepthGuard(t *testing.T) {
	deep := buildDeepList(t, maxExportDepth*3)
	obj, ok := deep[0].V.(PointerValue).Base.(Object)
	require.True(t, ok, "expected the list head's pointer base to be an Object")

	_, err := ExportObject(obj, 100_000_000)
	require.ErrorIs(t, err, ErrExportDepthExceeded)

	shallow := buildDeepList(t, 100)
	obj, ok = shallow[0].V.(PointerValue).Base.(Object)
	require.True(t, ok)
	_, err = ExportObject(obj, 100_000_000)
	require.NoError(t, err)
}

// buildDeepSlices evaluates a package that nests a []interface{} `depth` levels
// deep. Unlike buildDeepList this costs only ONE internal level per user level
// (the element TypedValue; the slice's base array is a peer hop, not counted),
// so it is the shape that buys the most user depth per byte charged — the one
// that sets the worst case in the ADR's residual arithmetic.
func buildDeepSlices(t *testing.T, depth int) []TypedValue {
	t.Helper()
	m := NewMachine("testdata", nil)
	t.Cleanup(m.Release)
	nn := m.MustParseFile("deep.gno", `package testdata
func Build(depth int) []interface{} {
	v := []interface{}{}
	for i := 0; i < depth; i++ {
		v = []interface{}{v}
	}
	return v
}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))
	return m.Eval(Call(Sel(Nx("testdata"), "Build"), X(depth)))
}

// TestExportValuesLimit_EffectiveUserDepth pins the boundary, and with it the
// ratio between maxExportDepth and the nesting a user actually writes — which is
// what the constant's doc comment and the ADR quote. The ratio is not 1:1 and
// not even fixed: it depends on the shape, so both ends are pinned here.
//
//   - pointer-linked: two internal levels per user level (the struct field, then
//     the heap item behind the pointer), so 500 nodes.
//   - slice-nested: one internal level per user level, so 1000 levels — twice
//     the depth for a comparable byte charge, which is why this end is the one
//     the residual arithmetic must use.
//
// If either ratio moves, the constant's doc comment and the ADR are wrong and
// must be updated with it. The accepted cases also marshal, since exporting a
// legitimate result is only useful if amino can then serialize it.
func TestExportValuesLimit_EffectiveUserDepth(t *testing.T) {
	for _, tc := range []struct {
		name    string
		build   func(*testing.T, int) []TypedValue
		wantMax int
	}{
		{"pointer-linked", buildDeepList, maxExportDepth / 2},
		{"slice-nested", buildDeepSlices, maxExportDepth},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exported, err := ExportValues(tc.build(t, tc.wantMax), 100_000_000)
			require.NoError(t, err, "%d levels should still export", tc.wantMax)
			bz, err := amino.MarshalJSON(exported)
			require.NoError(t, err)
			require.NotEmpty(t, bz)

			_, err = ExportValues(tc.build(t, tc.wantMax+1), 100_000_000)
			require.ErrorIs(t, err, ErrExportDepthExceeded,
				"%d levels should be rejected", tc.wantMax+1)
		})
	}
}
