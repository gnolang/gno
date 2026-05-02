package tests

import (
	"bytes"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

// registerLocal builds a fresh codec with the given types registered.
// Each test gets its own codec so registration-time panics are isolated.
func registerLocal(t *testing.T, types ...any) *amino.Codec {
	t.Helper()
	cdc := amino.NewCodec()
	p := pkg.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino/tests",
		"tests_varint",
		pkg.GetCallersDirname(),
	).WithTypes(types...)
	cdc.RegisterPackage(p)
	cdc.Seal()
	return cdc
}

// Phase 7 of GNOKMSAMINOVARINT.md: tests for the new binary:"varint" tag.
// These verify the new plain-varint encoding path matches upstream protobuf
// int64/int32 wire bytes, that mutual-exclusion validation fires at
// registration, and that the existing zigzag path is unaffected.

// ---- Phase 7.2 / 7.4: encoder primitives (no codec needed) -----------------

// Plain varint of -1 takes the full 10 bytes (sign-extended uint64
// 0xFFFFFFFFFFFFFFFF). Zigzag of -1 takes 1 byte. This is the load-bearing
// difference for upstream-Tendermint compatibility on POLRound and round/
// pol_round fields.
func TestPlainVarint_NegativeOneBytes(t *testing.T) {
	t.Parallel()

	var plain bytes.Buffer
	require.NoError(t, amino.EncodePlainVarint(&plain, -1))
	want := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}
	assert.Equal(t, want, plain.Bytes())
	assert.Equal(t, 10, plain.Len())

	var zigzag bytes.Buffer
	require.NoError(t, amino.EncodeVarint(&zigzag, -1))
	assert.Equal(t, []byte{0x01}, zigzag.Bytes())

	// Plain varint must match google.golang.org/protobuf/protowire's encoding
	// of int64(-1) wire-format. protowire.AppendVarint takes uint64; for proto
	// int64, callers cast int64 → uint64 directly.
	negOne := int64(-1)
	stdlib := protowire.AppendVarint(nil, uint64(negOne))
	assert.Equal(t, plain.Bytes(), stdlib)
}

// int32(-1) under plain varint produces the same 10 bytes as int64(-1),
// because protobuf int32 wire-encodes as if widened to int64.
func TestPlainVarint_Int32SignExtend(t *testing.T) {
	t.Parallel()

	var i32buf bytes.Buffer
	require.NoError(t, amino.EncodePlainVarint32(&i32buf, -1))

	var i64buf bytes.Buffer
	require.NoError(t, amino.EncodePlainVarint(&i64buf, -1))

	assert.Equal(t, i64buf.Bytes(), i32buf.Bytes())
	assert.Equal(t, 10, i32buf.Len())
}

// Sentinel byte tables for the most diagnostic values.
func TestPlainVarint_ByteTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		val  int64
		want []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"one", 1, []byte{0x01}},
		{"127", 127, []byte{0x7f}},
		{"128", 128, []byte{0x80, 0x01}},
		{"max-int32", math.MaxInt32, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
		// MinInt32 sign-extends in the upper 32 bits before varint.
		{"min-int32", math.MinInt32, []byte{0x80, 0x80, 0x80, 0x80, 0xf8, 0xff, 0xff, 0xff, 0xff, 0x01}},
		{"max-int64", math.MaxInt64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{"min-int64", math.MinInt64, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
		{"neg-one", -1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			require.NoError(t, amino.EncodePlainVarint(&buf, c.val))
			assert.Equal(t, c.want, buf.Bytes())
			assert.Equal(t, len(c.want), amino.PlainVarintSize(c.val))

			// Round-trip.
			got, n, err := amino.DecodePlainVarint(c.want)
			require.NoError(t, err)
			assert.Equal(t, len(c.want), n)
			assert.Equal(t, c.val, got)
		})
	}
}

func TestPlainVarint32_OverflowReject(t *testing.T) {
	t.Parallel()

	// Encode int64(MaxInt32 + 1) as plain varint, then try DecodePlainVarint32.
	var buf bytes.Buffer
	require.NoError(t, amino.EncodePlainVarint(&buf, int64(math.MaxInt32)+1))
	_, _, err := amino.DecodePlainVarint32(buf.Bytes())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "int32 overflow")
}

func TestPlainVarintReversed_MatchesForward(t *testing.T) {
	t.Parallel()

	// AppendPlainVarintReversed should produce the reversed byte order of
	// the forward-emit path, used by reverse-emit codegen sites.
	cases := []int64{0, 1, -1, math.MaxInt64, math.MinInt64}
	for _, v := range cases {
		var fwd bytes.Buffer
		require.NoError(t, amino.EncodePlainVarint(&fwd, v))
		rev := amino.AppendPlainVarintReversed(nil, v)

		// Reverse-emit reverses the byte sequence.
		fwdBytes := fwd.Bytes()
		expectedRev := make([]byte, len(fwdBytes))
		for i := range fwdBytes {
			expectedRev[i] = fwdBytes[len(fwdBytes)-1-i]
		}
		assert.Equal(t, expectedRev, rev, "AppendPlainVarintReversed(%d)", v)
	}
}

// ---- Phase 7.1 / 7.3 / 7.10: round-trip through codec ----------------------

type plainVarintStruct struct {
	A int64 `binary:"varint"`
	B int32 `binary:"varint"`
}

// zigzagStruct mirrors plainVarintStruct without the tag, for size comparison.
type zigzagStruct struct {
	A int64
	B int32
}

func TestPlainVarintCodec_RoundTrip(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, plainVarintStruct{})

	cases := []plainVarintStruct{
		{A: 0, B: 0},
		{A: 1, B: 1},
		{A: -1, B: -1},
		{A: math.MaxInt64, B: math.MaxInt32},
		{A: math.MinInt64, B: math.MinInt32},
	}
	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			bz, err := cdc.MarshalSized(&c)
			require.NoError(t, err)

			var got plainVarintStruct
			require.NoError(t, cdc.UnmarshalSized(bz, &got))
			assert.Equal(t, c, got)
		})
	}
}

// Negative values produce strictly more bytes under plain varint than zigzag.
func TestPlainVarintCodec_LargerThanZigzag(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, plainVarintStruct{}, zigzagStruct{})

	plain, err := cdc.MarshalSized(&plainVarintStruct{A: -1, B: -1})
	require.NoError(t, err)
	zigzag, err := cdc.MarshalSized(&zigzagStruct{A: -1, B: -1})
	require.NoError(t, err)
	assert.Greater(t, len(plain), len(zigzag),
		"plain varint of negative values must be longer than zigzag")
}

// ---- Phase 7.5 / 7.6 / 7.7: registration-time validation -------------------

// Both fixed64 and varint on the same field must panic at registration.
// This test only exercises the validation if Phase 1's comma-split parser
// is in place; without it, the entire string fails the switch and no flag
// gets set, so the test would silently no-op.
type bothFixedAndVarintStruct struct {
	X int64 `binary:"varint,fixed64"`
}

func TestPlainVarint_MutualExclusion(t *testing.T) {
	t.Parallel()
	// parseStructInfoWLocked recovers and re-panics with a generic
	// "panic parsing struct ..." wrapper, so we can't assert the inner
	// message; assert that registration panics at all.
	assert.Panics(t, func() { registerLocal(t, bothFixedAndVarintStruct{}) })
}

type bothFixed32AndFixed64Struct struct {
	X int64 `binary:"fixed32,fixed64"`
}

// TestFixed32AndFixed64_MutualExclusion: the mutex check must catch
// any pair of binary-encoding tags, not just varint+fixed*. Before
// the unified set>1 check, "fixed32,fixed64" silently set both flags
// and fell through to the fixed32 branch, panicking with a misleading
// "non-32bit type" message instead of flagging the real problem.
func TestFixed32AndFixed64_MutualExclusion(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() { registerLocal(t, bothFixed32AndFixed64Struct{}) })
}

type varintOnString struct {
	X string `binary:"varint"`
}

type varintOnBool struct {
	X bool `binary:"varint"`
}

type varintOnInt8 struct {
	X int8 `binary:"varint"`
}

type varintOnInt16 struct {
	X int16 `binary:"varint"`
}

func TestPlainVarint_RejectWrongTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		typ  any
	}{
		{"string", varintOnString{}},
		{"bool", varintOnBool{}},
		{"int8", varintOnInt8{}},
		{"int16", varintOnInt16{}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Panics(t, func() { registerLocal(t, c.typ) })
		})
	}
}

// ---- Phase 7.14: wire compat with stdlib protobuf --------------------------

// Encode an int64-tagged-varint field via amino, decode the inner field bytes
// with google.golang.org/protobuf/protowire (stdlib). This is the definitive
// interop test — bytes that pass this round-trip will work with any other
// protobuf-compatible decoder, including unmodified tmkms.
//
// Note: the amino-emitted bytes include framing (length prefix from
// MarshalSized + field-tag + payload). We extract just the payload by
// skipping the length prefix and field-tag.
func TestPlainVarint_StdlibProtobufInterop(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, plainVarintStruct{})

	// Use a value where plain and zigzag would differ.
	v := plainVarintStruct{A: -42, B: -42}
	bz, err := cdc.Marshal(&v) // unsized — no length prefix
	require.NoError(t, err)

	// Walk the encoded bytes, parsing each (field-num, wire-type) tag and
	// the corresponding varint value via protowire. For the int64 field
	// tagged binary:"varint" the wire-type must be 0 (Varint) and the
	// value must decode as an unsigned varint whose int64-cast equals -42.
	cur := bz
	gotFields := map[int]int64{}
	for len(cur) > 0 {
		num, typ, n := protowire.ConsumeTag(cur)
		require.Greater(t, n, 0)
		cur = cur[n:]
		require.Equal(t, protowire.VarintType, typ, "field %d wire-type", num)

		u, n := protowire.ConsumeVarint(cur)
		require.Greater(t, n, 0)
		cur = cur[n:]
		gotFields[int(num)] = int64(u)
	}

	// Both fields must have decoded to -42 under int64-cast-from-uint64
	// (the protobuf int64/int32 convention).
	require.Len(t, gotFields, 2)
	assert.Equal(t, int64(-42), gotFields[1])
	assert.Equal(t, int64(-42), gotFields[2])
}

// ---- Phase 7.15: regression — tag-less fields unaffected -------------------

// A struct with NO binary:"varint" tag must produce byte-identical output
// to before the patch. We can't compare against a frozen byte sequence here
// (since we just patched), but we CAN assert the encoding still uses zigzag
// — i.e., int64(-1) is 1 byte, not 10.
type tagLessStruct struct {
	A int64
	B int32
}

func TestPlainVarint_TagLessUnaffected(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, tagLessStruct{})

	bz, err := cdc.Marshal(&tagLessStruct{A: -1, B: -1})
	require.NoError(t, err)

	// Walk the bytes; each Varint field should consume exactly 1 byte for -1
	// under zigzag (sint64/sint32 → varint of 1 = 0x01).
	cur := bz
	count := 0
	for len(cur) > 0 {
		_, _, n := protowire.ConsumeTag(cur)
		require.Greater(t, n, 0)
		cur = cur[n:]

		u, n := protowire.ConsumeVarint(cur)
		require.Greater(t, n, 0)
		cur = cur[n:]
		// zigzag-decode: (u >> 1) ^ -(u & 1)
		v := int64(u>>1) ^ -int64(u&1)
		assert.Equal(t, int64(-1), v)
		assert.Equal(t, 1, n, "zigzag of -1 must be 1 byte (untouched by patch)")
		count++
	}
	assert.Equal(t, 2, count)
}

// ---- Phase 7.11 / 7.12 / 7.13: nested-list propagation --------------------
//
// The Phase 5b/5c fixes (`nListFieldOptions` preserving BinPlainVarint, and
// `NList.Name()` adding a `Varint` distinguisher) are only exercised through
// the schema-generator's nested-list path. Without these tests the regression
// surface is silent: a future edit to `nListFieldOptions` that drops
// BinPlainVarint would still encode bytes correctly per-element, but emit a
// schema that says `repeated sint64` while the bytes are plain-varint —
// schema/wire mismatch, stdlib-protobuf decoder rejects.

// listVarintField has a slice-of-int64 field tagged binary:"varint". This
// must be a struct field (not a top-level type) for findNLists to enter the
// struct branch where nListFieldOptions is consulted.
type listVarintField struct {
	Xs []int64 `binary:"varint"`
}

type listListVarintField struct {
	Xss [][]int64 `binary:"varint"`
}

// bothListsStruct has both a default (zigzag) and varint-tagged [][]int64
// in the same package, exercising NList type-name disambiguation. Note: the
// disambiguation only matters at depth >= 2, because single-level []int64
// emits inline `repeated <type>` and never produces a synthetic NList
// message (and so cannot collide). Doubly-nested produces synthetic types
// like `TESTS_Int64ValueList` (zigzag) vs `TESTS_VarintInt64ValueList` —
// these MUST have distinct names or the proto file is malformed.
type bothListsStruct struct {
	Zigzag [][]int64
	Plain  [][]int64 `binary:"varint"`
}

// schemaFor renders the generated proto schema text for a registered struct.
func schemaFor(t *testing.T, types ...any) string {
	t.Helper()
	p3c := genproto.NewP3Context()
	p := pkg.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino/tests",
		"tests_varint",
		pkg.GetCallersDirname(),
	).WithTypes(types...)
	p3c.RegisterPackage(p)
	rtz := make([]reflect.Type, len(types))
	for i, ty := range types {
		rtz[i] = reflect.TypeOf(ty)
	}
	doc := p3c.GenerateProto3SchemaForTypes(p, rtz...)
	return doc.Print()
}

// Test #11 — single-nested list: `[]int64 binary:"varint"` as a struct field
// must produce `repeated int64` (NOT `repeated sint64`) in the proto schema,
// AND emit plain-varint-encoded element bytes on the wire.
func TestPlainVarint_SingleNestedList_Schema(t *testing.T) {
	t.Parallel()
	schema := schemaFor(t, listVarintField{})
	t.Logf("schema:\n%s", schema)
	// The nested-list element type uses the field's varint encoding.
	assert.Contains(t, schema, "repeated int64",
		"nested list of int64 with binary:\"varint\" must emit `repeated int64`, not sint64")
	assert.NotContains(t, schema, "repeated sint64",
		"schema must NOT contain `repeated sint64` for the varint-tagged list")
}

func TestPlainVarint_SingleNestedList_Wire(t *testing.T) {
	t.Parallel()
	cdc := registerLocal(t, listVarintField{})
	bz, err := cdc.Marshal(&listVarintField{Xs: []int64{-1, 1}})
	require.NoError(t, err)

	// Skip the field-tag for the outer list.
	cur := bz
	_, typ, n := protowire.ConsumeTag(cur)
	require.Greater(t, n, 0)
	require.Equal(t, protowire.BytesType, typ, "packed list uses BytesType wire-type")
	cur = cur[n:]

	// Length-prefix of the packed list.
	listLen, n := protowire.ConsumeVarint(cur)
	require.Greater(t, n, 0)
	cur = cur[n:]
	require.Equal(t, int(listLen), len(cur), "list length-prefix must match remaining bytes")

	// First element: -1 must be 10 bytes plain varint.
	u1, n1 := protowire.ConsumeVarint(cur)
	require.Greater(t, n1, 0)
	cur = cur[n1:]
	assert.Equal(t, int64(-1), int64(u1))
	assert.Equal(t, 10, n1, "plain varint of -1 must be 10 bytes")

	// Second element: 1 must be 1 byte.
	u2, n2 := protowire.ConsumeVarint(cur)
	require.Greater(t, n2, 0)
	assert.Equal(t, int64(1), int64(u2))
	assert.Equal(t, 1, n2)
}

// Test #12 — doubly-nested list: `[][]int64 binary:"varint"`. The inner list
// element schema must still emit `int64` (not `sint64`). This exercises
// nListFieldOptions propagation through findNLists2 recursion at every
// depth.
func TestPlainVarint_DoublyNestedList_Schema(t *testing.T) {
	t.Parallel()
	schema := schemaFor(t, listListVarintField{})
	t.Logf("schema:\n%s", schema)
	// At least one synthetic NList message must contain `repeated int64`
	// (the inner list element). If nListFieldOptions stripped the flag,
	// the inner element would emit `sint64` and this would fail.
	assert.Contains(t, schema, "repeated int64",
		"doubly-nested list with binary:\"varint\" must propagate to inner element schema")
	assert.NotContains(t, schema, "repeated sint64",
		"schema must NOT contain `repeated sint64` anywhere")
}

// Test #13 — NList type-name disambiguation. A package containing both a
// default `[][]int64` and a varint-tagged `[][]int64` field must generate
// two distinct synthetic NList type names. Without the `Varint` prefix in
// `NList.Name()`, both NLists would collide on the same generated message
// name, producing a duplicate-message-name proto schema.
func TestPlainVarint_NListNameDisambiguation(t *testing.T) {
	t.Parallel()
	schema := schemaFor(t, bothListsStruct{})
	t.Logf("schema:\n%s", schema)

	// Both kinds of synthetic NList element type must be emitted. The
	// genproto package prefix is `TESTS_` (capitalized P3PkgName) and the
	// per-NList suffix follows from `NList.Name()`: the zigzag variant
	// gets just `Int64ValueList`, the varint variant gets `VarintInt64ValueList`.
	assert.Contains(t, schema, "TESTS_Int64ValueList",
		"default-zigzag [][]int64 must produce a synthetic NList type with `Int64ValueList` in the name")
	assert.Contains(t, schema, "TESTS_VarintInt64ValueList",
		"plain-varint [][]int64 must produce a distinct synthetic NList type prefixed with `Varint`")

	// Each synthetic message must appear exactly once. Without the
	// `Varint` distinguisher in NList.Name(), both NLists would emit the
	// same `message TESTS_Int64ValueList {` declaration twice — proto
	// error.
	for _, name := range []string{"TESTS_Int64ValueList", "TESTS_VarintInt64ValueList"} {
		needle := "message " + name + " {"
		count := strings.Count(schema, needle)
		assert.Equal(t, 1, count, "expected exactly one `%s` definition, found %d", needle, count)
	}
}

