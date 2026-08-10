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
// method values (Func == nil) re-emits the attacker-controlled method-name
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

// TestExportValuesLimit_String verifies that a large ephemeral string is
// rejected by the size bound before the export walk completes, while a small
// one passes.
func TestExportValuesLimit_String(t *testing.T) {
	tv := TypedValue{T: StringType, V: StringValue(strings.Repeat("A", 1_000_000))}

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
