package tests

// Tests for amino:"skip_zero": a tagged field is left out of the binary and
// JSON encodings whenever its Go value is zero, even where its repr would
// encode to bytes, and decodes back to that zero value.

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
)

// skipZeroCodec registers this package, whose FuzzSkipZero and friends
// carry amino:"skip_zero" and have genproto2 code.
func skipZeroCodec(t *testing.T) *amino.Codec {
	t.Helper()
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	return cdc
}

// skipZeroTwin is FuzzSkipZero without its amino:"skip_zero" tags, to show
// what the tags change. The two types convert to each other.
type skipZeroTwin struct {
	Addr       SkipZeroAddr
	Bytes      [4]byte
	Ints       [2]int64
	Inners     [2]SkipZeroInner
	Inner      SkipZeroInner
	List       SkipZeroList
	Score      float64 `amino:"unsafe"`
	Fixed      int64   `binary:"fixed64"`
	Name       string  `json:"name,omitempty"`
	Nums       []int32
	PlainAddr  SkipZeroAddr
	PlainInner SkipZeroInner
}

// fullSkipZero has every field non-zero.
func fullSkipZero() FuzzSkipZero {
	return FuzzSkipZero{
		Addr:       SkipZeroAddr{1, 2, 3, 4},
		Bytes:      [4]byte{5},
		Ints:       [2]int64{0, -6},
		Inners:     [2]SkipZeroInner{{}, {N: 7}},
		Inner:      SkipZeroInner{N: 8},
		List:       SkipZeroList{Ns: []int32{9}},
		Score:      1.5,
		Fixed:      -10,
		Name:       "eleven",
		Nums:       []int32{12},
		PlainAddr:  SkipZeroAddr{13},
		PlainInner: SkipZeroInner{N: 14},
	}
}

// zeroOneField zeroes each tagged field of FuzzSkipZero in turn, keyed by
// its field number.
var zeroOneField = []struct {
	num  protowire.Number
	zero func(*FuzzSkipZero)
}{
	{1, func(v *FuzzSkipZero) { v.Addr = SkipZeroAddr{} }},
	{2, func(v *FuzzSkipZero) { v.Bytes = [4]byte{} }},
	{3, func(v *FuzzSkipZero) { v.Ints = [2]int64{} }},
	{4, func(v *FuzzSkipZero) { v.Inners = [2]SkipZeroInner{} }},
	{5, func(v *FuzzSkipZero) { v.Inner = SkipZeroInner{} }},
	{6, func(v *FuzzSkipZero) { v.List = SkipZeroList{} }},
	{7, func(v *FuzzSkipZero) { v.Score = 0 }},
	{8, func(v *FuzzSkipZero) { v.Fixed = 0 }},
	{9, func(v *FuzzSkipZero) { v.Name = "" }},
	{10, func(v *FuzzSkipZero) { v.Nums = nil }},
}

// wireFieldNums lists the field numbers in a struct's binary encoding, in
// order, once per occurrence.
func wireFieldNums(t *testing.T, bz []byte) []protowire.Number {
	t.Helper()
	var nums []protowire.Number
	for len(bz) > 0 {
		num, _, n := protowire.ConsumeField(bz)
		require.GreaterOrEqual(t, n, 0, "malformed encoding: %X", bz)
		nums = append(nums, num)
		bz = bz[n:]
	}
	return nums
}

type skipZeroWriteEmpty struct {
	A string `amino:"skip_zero,write_empty"`
}

type skipZeroPointer struct {
	P *int64 `amino:"skip_zero"`
}

type skipZeroTime struct {
	T time.Time `amino:"skip_zero"`
}

type skipZeroTimePointer struct {
	T *time.Time `amino:"skip_zero"`
}

// TestSkipZeroRefused: registration refuses amino:"skip_zero" next to
// write_empty, and wherever a missing field decodes to something other than
// the zero value the tag would leave out.
func TestSkipZeroRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		v    any
		want string
	}{
		{skipZeroWriteEmpty{}, `amino:"skip_zero" and amino:"write_empty" are mutually exclusive`},
		{skipZeroPointer{}, `amino:"skip_zero" is not allowed on a *int64 field`},
		{skipZeroTime{}, `amino:"skip_zero" is not allowed on a time.Time field`},
		{skipZeroTimePointer{}, `amino:"skip_zero" is not allowed on a *time.Time field`},
	} {
		t.Run(reflect.TypeOf(tc.v).Name(), func(t *testing.T) {
			t.Parallel()
			defer func() {
				r := recover()
				require.NotNil(t, r, "registration should panic")
				assert.Contains(t, fmt.Sprint(r), tc.want)
			}()
			registerLocal(t, tc.v)
		})
	}
}

type skipZeroDuration struct {
	D time.Duration `amino:"skip_zero"`
}

type skipZeroStructPointer struct {
	P *GnoVMPos `amino:"skip_zero"`
}

type skipZeroDropped struct {
	A string `json:"-" amino:"skip_zero,write_empty"`
}

// TestSkipZeroAllowed: the tag is fine wherever a missing field decodes to
// zero, as on a time.Duration and on a pointer to a struct, whose nil it
// leaves out of JSON too, and on a field json:"-" drops before reading tags.
func TestSkipZeroAllowed(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, skipZeroDuration{}, skipZeroStructPointer{}, skipZeroDropped{})
	for _, d := range []time.Duration{0, -1, time.Second} {
		bz, err := cdc.MarshalReflect(skipZeroDuration{D: d})
		require.NoError(t, err)
		assert.Equal(t, d == 0, len(bz) == 0, "duration %v encoded as %X", d, bz)
		var got skipZeroDuration
		require.NoError(t, cdc.UnmarshalReflect(bz, &got))
		assert.Equal(t, d, got.D)
	}

	for _, v := range []skipZeroStructPointer{{}, {P: &GnoVMPos{}}, {P: &GnoVMPos{Line: 1}}} {
		bz, err := cdc.MarshalReflect(v)
		require.NoError(t, err)
		assert.Equal(t, v.P == nil, len(bz) == 0, "%+v encoded as %X", v, bz)
		var got skipZeroStructPointer
		require.NoError(t, cdc.UnmarshalReflect(bz, &got))
		assert.Equal(t, v, got)

		js, err := cdc.JSONMarshal(v)
		require.NoError(t, err)
		assert.Equal(t, v.P == nil, string(js) == "{}", "%+v in JSON: %s", v, js)
		var gotJSON skipZeroStructPointer
		require.NoError(t, cdc.JSONUnmarshal(js, &gotJSON))
		assert.Equal(t, v, gotJSON)
	}

	bz, err := cdc.MarshalReflect(skipZeroDropped{A: "dropped"})
	require.NoError(t, err)
	assert.Empty(t, bz)
}

// TestSkipZeroBinary: a tagged field is left out exactly when its Go value
// is zero, including where its repr would encode to bytes, and is written
// exactly as without the tag otherwise.
func TestSkipZeroBinary(t *testing.T) {
	t.Parallel()

	cdc := skipZeroCodec(t)
	twinCdc := registerLocal(t, skipZeroTwin{})

	// All zero: only the untagged controls, fields 11 and 12, are written.
	// Untagged, the zero Addr, Bytes, Ints, Inners, Inner, List and Score
	// would all write bytes; Inners writes one entry per element.
	bz, err := cdc.MarshalReflect(FuzzSkipZero{})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{11, 12}, wireFieldNums(t, bz))
	twinBz, err := twinCdc.MarshalReflect(skipZeroTwin{})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{1, 2, 3, 4, 4, 5, 6, 7, 11, 12}, wireFieldNums(t, twinBz))

	// All non-zero: byte for byte the encoding without the tags.
	bz, err = cdc.MarshalReflect(fullSkipZero())
	require.NoError(t, err)
	twinBz, err = twinCdc.MarshalReflect(skipZeroTwin(fullSkipZero()))
	require.NoError(t, err)
	assert.Equal(t, twinBz, bz)

	// Zeroing one tagged field leaves out that field and nothing else.
	fullNums := wireFieldNums(t, twinBz)
	for _, tc := range zeroOneField {
		v := fullSkipZero()
		tc.zero(&v)
		bz, err := cdc.MarshalReflect(v)
		require.NoError(t, err)
		want := slices.DeleteFunc(slices.Clone(fullNums), func(n protowire.Number) bool { return n == tc.num })
		assert.Equal(t, want, wireFieldNums(t, bz), "zeroed field %d", tc.num)
	}

	// Values that only look zero stay on the wire: a struct holding a
	// non-nil empty slice, and an array with one zero element.
	bz, err = cdc.MarshalReflect(FuzzSkipZero{
		Inners: [2]SkipZeroInner{{N: 1}, {}},
		List:   SkipZeroList{Ns: []int32{}},
	})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{4, 4, 6, 11, 12}, wireFieldNums(t, bz))

	// Zero is what == says, as reflect.Value.IsZero has agreed since Go
	// 1.22: a float -0.0 is zero, so it is left out and decodes as 0, while
	// a NaN is not.
	bz, err = cdc.MarshalReflect(FuzzSkipZero{Score: math.Copysign(0, -1)})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{11, 12}, wireFieldNums(t, bz))
	var got FuzzSkipZero
	require.NoError(t, cdc.UnmarshalReflect(bz, &got))
	assert.False(t, math.Signbit(got.Score))
	bz, err = cdc.MarshalReflect(FuzzSkipZero{Score: math.NaN()})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{7, 11, 12}, wireFieldNums(t, bz))

	// An interface is zero only when nil; one holding a zero value is
	// written.
	bz, err = cdc.MarshalReflect(SkipZeroInterface{})
	require.NoError(t, err)
	assert.Empty(t, bz)
	bz, err = cdc.MarshalReflect(SkipZeroInterface{I: Concrete1{}})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{1}, wireFieldNums(t, bz))

	// Next to nil_elements: the all-nil array is the zero value and is left
	// out; an array with one element set writes both, the nil one as
	// nil_elements does.
	bz, err = cdc.MarshalReflect(SkipZeroNilElements{})
	require.NoError(t, err)
	assert.Empty(t, bz)
	bz, err = cdc.MarshalReflect(SkipZeroNilElements{Array: [2]*GnoVMPos{nil, {Line: 1}}})
	require.NoError(t, err)
	assert.Equal(t, []protowire.Number{2, 2}, wireFieldNums(t, bz))

	// A struct whose fields are all left out is empty, so an untagged field
	// of its type is left out in turn.
	bz, err = cdc.MarshalReflect(SkipZeroHolder{})
	require.NoError(t, err)
	assert.Empty(t, bz)
}

// TestSkipZeroDecodesIntoUsedValue: a field left out decodes to zero even
// into a value that held something, through both decoders. Inners, an
// unpacked list, absent before a field that is present, is the case the
// reflect decoder used to leave as it found it.
func TestSkipZeroDecodesIntoUsedValue(t *testing.T) {
	t.Parallel()

	cdc := skipZeroCodec(t)
	want := FuzzSkipZero{Inner: SkipZeroInner{N: 3}}
	bz, err := cdc.MarshalReflect(want)
	require.NoError(t, err)
	require.Equal(t, []protowire.Number{5, 11, 12}, wireFieldNums(t, bz))

	viaReflect, viaBinary2 := fullSkipZero(), fullSkipZero()
	require.NoError(t, cdc.UnmarshalReflect(bz, &viaReflect))
	require.NoError(t, viaBinary2.UnmarshalBinary2(cdc, bz, 0))
	assert.Equal(t, want, viaReflect)
	assert.Equal(t, want, viaBinary2)
}

// TestSkipZeroParity: the genproto2 code agrees with the reflect codec on
// every byte, on SizeBinary2, and on the decoded value.
func TestSkipZeroParity(t *testing.T) {
	t.Parallel()

	cdc := skipZeroCodec(t)
	full := fullSkipZero()
	cases := []any{
		&FuzzSkipZero{},
		&full,
		&FuzzSkipZero{Score: math.Copysign(0, -1)},
		&FuzzSkipZero{Inners: [2]SkipZeroInner{{N: 1}, {}}},
		&FuzzSkipZero{Inners: [2]SkipZeroInner{{}, {Addr: SkipZeroAddr{1}}}},
		&FuzzSkipZero{List: SkipZeroList{Addr: SkipZeroAddr{2}}},
		&SkipZeroInner{},
		&SkipZeroList{},
		&SkipZeroInterface{},
		&SkipZeroInterface{I: Concrete1{}},
		&SkipZeroNilElements{},
		&SkipZeroNilElements{
			Slice: []*GnoVMPos{nil, {Line: 1}},
			Array: [2]*GnoVMPos{nil, {Line: 2}},
		},
		&SkipZeroHolder{},
		&SkipZeroHolder{N: 1},
		&SkipZeroHolder{Only: SkipZeroOnly{Addr: SkipZeroAddr{1}}},
		// The elements' tagged fields go through genproto2's inline writer.
		&SkipZeroElems{},
		&SkipZeroElems{Items: []SkipZeroElem{
			{},
			{N: 1},
			{Arr: [2]int32{0, 2}},
			{Inner: SkipZeroInner{N: 3}},
			{Inner: SkipZeroInner{Addr: SkipZeroAddr{4}}},
		}},
	}
	for _, tc := range zeroOneField {
		v := fullSkipZero()
		tc.zero(&v)
		cases = append(cases, &v)
	}
	for i, v := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, v)
		})
	}

	// A non-nil empty slice makes List non-zero although it encodes to
	// nothing itself, and it decodes back as nil, so the round trip is not
	// exact here; compare the codecs with each other only.
	compareEncoding(t, cdc, "List/empty-slice", FuzzSkipZero{List: SkipZeroList{Ns: []int32{}}})
}

// TestSkipZeroReadsEncodingsWrittenWithoutIt: bytes written before a field
// had the tag hold its zero value explicitly. Both decoders still read
// them, to the same value the new, shorter encoding gives. This is what
// lets an existing field take the tag without breaking stored data.
func TestSkipZeroReadsEncodingsWrittenWithoutIt(t *testing.T) {
	t.Parallel()

	cdc := skipZeroCodec(t)
	twinCdc := registerLocal(t, skipZeroTwin{})
	for i, v := range []FuzzSkipZero{{}, {Inner: SkipZeroInner{N: 3}}, fullSkipZero()} {
		old, err := twinCdc.MarshalReflect(skipZeroTwin(v))
		require.NoError(t, err)

		var viaReflect, viaBinary2 FuzzSkipZero
		require.NoError(t, cdc.UnmarshalReflect(old, &viaReflect))
		require.NoError(t, viaBinary2.UnmarshalBinary2(cdc, old, 0))
		assert.Equal(t, v, viaReflect, "case %d", i)
		assert.Equal(t, v, viaBinary2, "case %d", i)

		now, err := cdc.MarshalReflect(v)
		require.NoError(t, err)
		again, err := cdc.MarshalReflect(viaReflect)
		require.NoError(t, err)
		assert.Equal(t, now, again, "case %d", i)
	}
}

// TestSkipZeroJSON: the tag leaves zero values out of JSON as well, and a
// missing tagged field decodes to zero.
func TestSkipZeroJSON(t *testing.T) {
	t.Parallel()

	cdc := skipZeroCodec(t)
	twinCdc := registerLocal(t, skipZeroTwin{})
	keys := func(bz []byte) []string {
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(bz, &m))
		return slices.Sorted(maps.Keys(m))
	}

	// All zero: only the untagged controls are written. Untagged, every
	// field but the omitempty name would be.
	zeroJSON, err := cdc.JSONMarshal(FuzzSkipZero{})
	require.NoError(t, err)
	assert.Equal(t, []string{"PlainAddr", "PlainInner"}, keys(zeroJSON))
	twinJSON, err := twinCdc.JSONMarshal(skipZeroTwin{})
	require.NoError(t, err)
	assert.Len(t, keys(twinJSON), 11)

	// All non-zero: the same JSON as without the tags.
	bz, err := cdc.JSONMarshal(fullSkipZero())
	require.NoError(t, err)
	twinBz, err := twinCdc.JSONMarshal(skipZeroTwin(fullSkipZero()))
	require.NoError(t, err)
	assert.Equal(t, string(twinBz), string(bz))

	// Zeroing one tagged field leaves out that key alone.
	for _, tc := range zeroOneField {
		v := fullSkipZero()
		tc.zero(&v)
		bz, err := cdc.JSONMarshal(v)
		require.NoError(t, err)
		assert.Len(t, keys(bz), 11, "zeroed field %d", tc.num)
	}

	// Missing tagged fields decode to zero even over a non-zero value, and
	// even next to omitempty, which alone leaves the old value in place.
	got := fullSkipZero()
	require.NoError(t, cdc.JSONUnmarshal(zeroJSON, &got))
	assert.Equal(t, FuzzSkipZero{}, got)
	twinGot := skipZeroTwin(fullSkipZero())
	require.NoError(t, twinCdc.JSONUnmarshal(twinJSON, &twinGot))
	assert.Equal(t, "eleven", twinGot.Name)

	// JSON written without the tags, zero values spelled out, still reads.
	got = fullSkipZero()
	require.NoError(t, cdc.JSONUnmarshal(twinJSON, &got))
	assert.Equal(t, FuzzSkipZero{}, got)
}
