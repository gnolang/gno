package amino_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

func TestNilSliceEmptySlice(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	type TestStruct struct {
		A []byte
		B []int
		C [][]byte
		D [][]int
		E []*[]byte
		F []*[]int
	}
	nnb, nni := []byte(nil), []int(nil)
	eeb, eei := []byte{}, []int{}

	a := TestStruct{
		A: nnb,
		B: nni,
		C: [][]byte{nnb},
		D: [][]int{nni},
		E: []*[]byte{nil},
		F: []*[]int{nil},
	}
	b := TestStruct{
		A: eeb,
		B: eei,
		C: [][]byte{eeb},
		D: [][]int{eei},
		E: []*[]byte{&nnb},
		F: []*[]int{&nni},
	}
	c := TestStruct{
		A: eeb,
		B: eei,
		C: [][]byte{eeb},
		D: [][]int{eei},
		E: []*[]byte{&eeb},
		F: []*[]int{&eei},
	}

	abz := cdc.MustMarshalSized(a)
	bbz := cdc.MustMarshalSized(b)
	cbz := cdc.MustMarshalSized(c)

	assert.Equal(t, abz, bbz, "a != b")
	assert.Equal(t, abz, cbz, "a != c")
}

func TestNewFieldBackwardsCompatibility(t *testing.T) {
	t.Parallel()

	type V1 struct {
		String  string
		String2 string
	}

	type V2 struct {
		String  string
		String2 string
		// new fields in V2:
		Time time.Time
		Int  int
	}

	type SomeStruct struct {
		Sth int
	}

	type V3 struct {
		String string
		// different from V1 starting here:
		Int  int
		Some SomeStruct
	}

	cdc := amino.NewCodec()
	notNow, err := time.Parse("2006-01-02", "1934-11-09")
	assert.NoError(t, err)
	v2 := V2{String: "hi", String2: "cosmos", Time: notNow, Int: 4}
	bz, err := cdc.Marshal(v2)
	assert.Nil(t, err, "unexpected error while encoding V2: %v", err)

	var v1 V1
	err = cdc.Unmarshal(bz, &v1)
	assert.Nil(t, err, "unexpected error %v", err)
	assert.Equal(t, v1, V1{"hi", "cosmos"},
		"backwards compatibility failed: didn't yield expected result ...")

	v3 := V3{String: "tender", Int: 2014, Some: SomeStruct{Sth: 84}}
	bz2, err := cdc.Marshal(v3)
	assert.Nil(t, err, "unexpected error")

	err = cdc.Unmarshal(bz2, &v1)
	// this might change later but we include this case to document the current behaviour:
	assert.NotNil(t, err, "expected an error here because of changed order of fields")

	// we still expect that decoding worked to some extend (until above error occurred):
	assert.Equal(t, v1, V1{"tender", "cosmos"})
}

func TestWriteEmpty(t *testing.T) {
	t.Parallel()

	type Inner struct {
		Val int
	}
	type SomeStruct struct {
		Inner Inner
	}

	cdc := amino.NewCodec()
	b, err := cdc.Marshal(Inner{})
	assert.NoError(t, err)
	assert.Equal(t, b, []byte(nil), "empty struct should be encoded as empty bytes")
	var inner Inner
	err = cdc.Unmarshal(b, &inner)
	require.NoError(t, err)
	assert.Equal(t, Inner{}, inner, "")

	b, err = cdc.Marshal(SomeStruct{})
	assert.NoError(t, err)
	assert.Equal(t, b, []byte(nil), "empty structs should be encoded as empty bytes")
	var outer SomeStruct
	err = cdc.Unmarshal(b, &outer)
	require.NoError(t, err)

	assert.Equal(t, SomeStruct{}, outer, "")
}

func TestForceWriteEmpty(t *testing.T) {
	t.Parallel()

	type InnerWriteEmpty struct {
		// sth. that isn't zero-len if default, e.g. fixed32:
		ValIn int32 `amino:"write_empty" binary:"fixed32"`
	}

	type OuterWriteEmpty struct {
		In  InnerWriteEmpty `amino:"write_empty"`
		Val int32           `amino:"write_empty" binary:"fixed32"`
	}

	cdc := amino.NewCodec()

	b, err := cdc.Marshal(OuterWriteEmpty{})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xa, 0x5, 0xd, 0x0, 0x0, 0x0, 0x0, 0x15, 0x0, 0x0, 0x0, 0x0}, b)

	b, err = cdc.Marshal(InnerWriteEmpty{})
	assert.NoError(t, err)
	assert.Equal(t, []byte{13, 0, 0, 0, 0}, b)
}

func TestStructSlice(t *testing.T) {
	t.Parallel()

	type Foo struct {
		A uint
		B uint
	}

	type Foos struct {
		Fs []Foo
	}

	f := Foos{Fs: []Foo{{100, 101}, {102, 103}}}

	cdc := amino.NewCodec()

	bz, err := cdc.Marshal(f)
	assert.NoError(t, err)
	assert.Equal(t, "0A04086410650A0408661067", fmt.Sprintf("%X", bz))
	t.Log(bz)
	var f2 Foos
	err = cdc.Unmarshal(bz, &f2)
	require.NoError(t, err)
	assert.Equal(t, f, f2)
}

func TestStructPointerSlice1(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	type Foo struct {
		A string
		B int
		C []*Foo `amino:"nil_elements"`
		D string // exposed
	}

	f := Foo{
		A: "k",
		B: 2,
		C: []*Foo{nil, nil, nil},
		D: "j",
	}
	bz, err := cdc.MarshalSized(f)
	assert.NoError(t, err)

	var f2 Foo
	err = cdc.UnmarshalSized(bz, &f2)
	assert.Nil(t, err)

	assert.Equal(t, f, f2)
	assert.Nil(t, f2.C[0])

	f3 := Foo{
		A: "k",
		B: 2,
		C: []*Foo{{}, {}, {}},
		D: "j",
	}
	bz2, err := cdc.MarshalSized(f3)
	assert.NoError(t, err)
	assert.Equal(t, bz, bz2, "empty slice elements should be encoded the same as nil")
}

// Like TestStructPointerSlice2, but without nil_elements field tag.
func TestStructPointerSlice2(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	type Foo struct {
		A string
		B int
		C []*Foo
		D string // exposed
	}

	f := Foo{
		A: "k",
		B: 2,
		C: []*Foo{nil, nil, nil},
		D: "j",
	}
	_, err := cdc.MarshalSized(f)
	assert.Error(t, err, "nil elements of a slice/array not supported unless nil_elements field tag set.")

	f.C = []*Foo{{}, {}, {}}
	bz, err := cdc.MarshalSized(f)
	assert.NoError(t, err)

	var f2 Foo
	err = cdc.UnmarshalSized(bz, &f2)
	assert.Nil(t, err)

	assert.Equal(t, f, f2)
	assert.NotNil(t, f2.C[0])
}

func TestBasicTypes(t *testing.T) {
	t.Parallel()

	// we explicitly disallow type definitions like the following:
	type byteAlias []byte

	cdc := amino.NewCodec()
	ba := byteAlias([]byte("this should work because it gets wrapped by a struct"))
	bz, err := cdc.MarshalSized(ba)
	assert.NotZero(t, bz)
	require.NoError(t, err)

	res := &byteAlias{}
	err = cdc.UnmarshalSized(bz, res)

	require.NoError(t, err)
	assert.Equal(t, ba, *res)
}

func TestUnmarshalMapBinary(t *testing.T) {
	t.Parallel()

	obj := new(map[string]int)
	cdc := amino.NewCodec()

	// Binary doesn't support decoding to a map...
	binBytes := []byte(`dontcare`)
	assert.Panics(t, func() {
		err := cdc.Unmarshal(binBytes, &obj)
		assert.Fail(t, "should have panicked but got err: %v", err)
	})

	assert.Panics(t, func() {
		err := cdc.Unmarshal(binBytes, obj)
		require.Error(t, err)
	})

	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.Marshal(obj)
		assert.Fail(t, "should have panicked but got bz: %X err: %v", bz, err)
	})
}

func TestUnmarshalFuncBinary(t *testing.T) {
	t.Parallel()

	obj := func() {}
	cdc := amino.NewCodec()
	// Binary doesn't support decoding to a func...
	binBytes := []byte(`dontcare`)
	err := cdc.UnmarshalSized(binBytes, &obj)
	// on length prefixed we return an error:
	assert.Error(t, err)

	assert.Panics(t, func() {
		err = cdc.Unmarshal(binBytes, &obj)
		require.Error(t, err)
	})

	err = cdc.Unmarshal(binBytes, obj)
	require.Error(t, err)
	require.Equal(t, err, amino.ErrNoPointer)

	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.MarshalSized(obj)
		assert.Fail(t, "should have panicked but got bz: %X err: %v", bz, err)
	})
}

func TestDuration(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	d0 := time.Duration(0)
	bz := cdc.MustMarshal(d0)
	assert.Equal(t, bz, []byte(nil))
	var d time.Duration
	var dPtr *time.Duration
	var dZero time.Duration
	err := cdc.Unmarshal(nil, &d)
	assert.NoError(t, err)
	assert.Equal(t, d, time.Duration(0))
	err = cdc.Unmarshal(nil, &dPtr)
	assert.NoError(t, err)
	assert.Equal(t, dPtr, &dZero)
}

func TestInterfaceTypeAssignability(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)

	// Wrap PrimitivesStruct in an `any` field to get Any/typeURL encoding.
	// Then try to unmarshal into a struct with Interface1 field.
	// PrimitivesStruct doesn't implement Interface1, so this should error.
	type AnyWrapper struct {
		Value any
	}
	type Interface1Wrapper struct {
		Value tests.Interface1
	}

	src := AnyWrapper{Value: tests.PrimitivesStruct{Int: 42}}
	bz, err := cdc.Marshal(src)
	require.NoError(t, err)

	var dst Interface1Wrapper
	err = cdc.Unmarshal(bz, &dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not assignable")
}

func TestInterfaceTypeAssignabilityOnDecodeError(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)

	// Same as TestInterfaceTypeAssignability but with corrupted bytes
	// to trigger the error path in decodeReflectBinaryAny where
	// rv.Set(irvSet) is called for debugging purposes.
	type AnyWrapper struct {
		Value any
	}
	type Interface1Wrapper struct {
		Value tests.Interface1
	}

	src := AnyWrapper{Value: tests.PrimitivesStruct{Int: 42}}
	bz, err := cdc.Marshal(src)
	require.NoError(t, err)

	// Corrupt some bytes to cause decode error (but keep typeURL intact)
	if len(bz) > 20 {
		bz[len(bz)-1] ^= 0xFF
		bz[len(bz)-2] ^= 0xFF
	}

	var dst Interface1Wrapper
	err = cdc.Unmarshal(bz, &dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not assignable")
}
