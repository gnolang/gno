package amino_test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

type SimpleStruct struct {
	String string
	Bytes  []byte
	Time   time.Time
}

func newSimpleStruct() SimpleStruct {
	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}
	return s
}

func TestMarshalUnmarshalPointer0(t *testing.T) {
	t.Parallel()

	s := newSimpleStruct()
	cdc := amino.NewCodec()
	b, err := cdc.MarshalSized(s) // no indirection
	assert.NoError(t, err)

	var s2 SimpleStruct
	err = cdc.UnmarshalSized(b, &s2) // no indirection
	assert.NoError(t, err)
	assert.Equal(t, s, s2)
}

func TestMarshalUnmarshalPointer1(t *testing.T) {
	t.Parallel()

	s := newSimpleStruct()
	cdc := amino.NewCodec()
	b, err := cdc.MarshalSized(&s) // extra indirection
	assert.NoError(t, err)

	var s2 SimpleStruct
	err = cdc.UnmarshalSized(b, &s2) // no indirection
	assert.NoError(t, err)
	assert.Equal(t, s, s2)
}

func TestMarshalUnmarshalPointer2(t *testing.T) {
	t.Parallel()

	s := newSimpleStruct()
	ptr := &s
	cdc := amino.NewCodec()
	assert.Panics(t, func() {
		cdc.MarshalSized(&ptr) // double extra indirection panics.
		cdc.RegisterPackage(tests.Package)
	})
}

func TestMarshalUnmarshalPointer3(t *testing.T) {
	t.Parallel()

	s := newSimpleStruct()
	cdc := amino.NewCodec()
	b, err := cdc.MarshalSized(s) // no indirection
	assert.NoError(t, err)

	var s2 *SimpleStruct
	err = cdc.UnmarshalSized(b, &s2) // extra indirection
	assert.NoError(t, err)
	assert.Equal(t, s, *s2)
}

func TestDecodeVarint8(t *testing.T) {
	t.Parallel()

	// DecodeVarint8 uses binary.Varint so we need to make
	// sure that all the values out of the range of [-128, 127]
	// return an error.
	tests := []struct {
		in      int64
		wantErr string
		want    int8
	}{
		{in: 0x7F, want: 0x7F},
		{in: -0x7F, want: -0x7F},
		{in: -0x80, want: -0x80},
		{in: 0x10, want: 0x10},

		{in: -0xFF, wantErr: "decoding int8"},
		{in: 0xFF, wantErr: "decoding int8"},
		{in: 0x100, wantErr: "decoding int8"},
		{in: -0x100, wantErr: "decoding int8"},
	}

	buf := make([]byte, 10)
	for i, tt := range tests {
		n := binary.PutVarint(buf, tt.in)
		gotI8, gotN, err := amino.DecodeVarint8(buf[:n])
		if tt.wantErr != "" {
			if err == nil {
				t.Errorf("#%d expected error=%q", i, tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d\ngotErr=%q\nwantSegment=%q", i, err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d unexpected error: %v", i, err)
			continue
		}

		if wantI8 := tt.want; gotI8 != wantI8 {
			t.Errorf("#%d gotI8=%d wantI8=%d", i, gotI8, wantI8)
		}
		if wantN := n; gotN != wantN {
			t.Errorf("#%d gotN=%d wantN=%d", i, gotN, wantN)
		}
	}
}

func TestDecodeVarint16(t *testing.T) {
	t.Parallel()

	// DecodeVarint16 uses binary.Varint so we need to make
	// sure that all the values out of the range of [-32768, 32767]
	// return an error.
	tests := []struct {
		in      int64
		wantErr string
		want    int16
	}{
		{in: -0x8000, want: -0x8000},
		{in: -0x7FFF, want: -0x7FFF},
		{in: -0x7F, want: -0x7F},
		{in: -0x80, want: -0x80},
		{in: 0x10, want: 0x10},

		{in: -0xFFFF, wantErr: "decoding int16"},
		{in: 0xFFFF, wantErr: "decoding int16"},
		{in: 0x10000, wantErr: "decoding int16"},
		{in: -0x10000, wantErr: "decoding int16"},
	}

	buf := make([]byte, 10)
	for i, tt := range tests {
		n := binary.PutVarint(buf, tt.in)
		gotI16, gotN, err := amino.DecodeVarint16(buf[:n])
		if tt.wantErr != "" {
			if err == nil {
				t.Errorf("#%d in=(%X) expected error=%q", i, tt.in, tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d\ngotErr=%q\nwantSegment=%q", i, err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d unexpected error: %v", i, err)
			continue
		}

		if wantI16 := tt.want; gotI16 != wantI16 {
			t.Errorf("#%d gotI16=%d wantI16=%d", i, gotI16, wantI16)
		}
		if wantN := n; gotN != wantN {
			t.Errorf("#%d gotN=%d wantN=%d", i, gotN, wantN)
		}
	}
}

func TestEncodeDecodeString(t *testing.T) {
	t.Parallel()

	s := "🔌🎉⛵︎♠️⎍"
	bs := []byte(s)
	di := len(bs) * 3 / 4
	b1 := bs[:di]
	b2 := bs[di:]

	// Encoding phase
	buf1 := new(bytes.Buffer)
	if err := amino.EncodeByteSlice(buf1, b1); err != nil {
		t.Fatalf("EncodeByteSlice(b1) = %v", err)
	}
	buf2 := new(bytes.Buffer)
	if err := amino.EncodeByteSlice(buf2, b2); err != nil {
		t.Fatalf("EncodeByteSlice(b2) = %v", err)
	}

	// Decoding phase
	e1 := buf1.Bytes()
	dec1, n, err := amino.DecodeByteSlice(e1)
	if err != nil {
		t.Errorf("DecodeByteSlice(e1) = %v", err)
	}
	if g, w := n, len(e1); g != w {
		t.Errorf("e1: length:: got = %d want = %d", g, w)
	}
	e2 := buf2.Bytes()
	dec2, n, err := amino.DecodeByteSlice(e2)
	if err != nil {
		t.Errorf("DecodeByteSlice(e2) = %v", err)
	}
	if g, w := n, len(e2); g != w {
		t.Errorf("e2: length:: got = %d want = %d", g, w)
	}
	joined := bytes.Join([][]byte{dec1, dec2}, []byte(""))
	if !bytes.Equal(joined, bs) {
		t.Errorf("got joined=(% X) want=(% X)", joined, bs)
	}
	js := string(joined)
	if js != s {
		t.Errorf("got string=%q want=%q", js, s)
	}
}

func TestCodecSeal(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.Seal()

	assert.Panics(t, func() { cdc.RegisterPackage(tests.Package) })
}

func TestDupTypesMustPanic(t *testing.T) {
	// duplicate types must panic
	t.Parallel()

	pkg := amino.NewPackage(
		reflect.TypeOf(SimpleStruct{}).PkgPath(),
		"amino_test",
		amino.GetCallersDirname(),
	)
	assert.PanicsWithError(
		t,
		"type amino_test.SimpleStruct already registered with package",
		func() {
			pkg.WithTypes(
				SimpleStruct{},
				SimpleStruct{},
			)
		})
}

func TestTypesOutsidePackageMustPanic(t *testing.T) {
	// adding concrete types from within a different package must panic
	// (use dependency instead)
	t.Parallel()

	makepkg := func() *amino.Package {
		return amino.NewPackage(
			reflect.TypeOf(tests.EmptyStruct{}).PkgPath(),
			"amino_test",
			amino.GetCallersDirname(),
		)
	}

	makepkg().WithTypes(tests.PrimitivesStruct{}) // from same package ✓

	assert.Panics(t, func() {
		makepkg().WithTypes(
			SimpleStruct{}, // from another package ✗
		)
	})
}

func TestDupNamesMustPanic(t *testing.T) {
	// adding types with the same names must panic
	t.Parallel()

	makepkg := func() *amino.Package {
		return amino.NewPackage(
			reflect.TypeOf(tests.EmptyStruct{}).PkgPath(),
			"amino_test",
			amino.GetCallersDirname(),
		)
	}
	makepkg().WithTypes(
		tests.EmptyStruct{}, "A",
		tests.PrimitivesStruct{}, "B",
		tests.ShortArraysStruct{}, "C",
	)
	assert.Panics(t, func() {
		makepkg().WithTypes(
			tests.EmptyStruct{}, "A",
			tests.PrimitivesStruct{}, "B",
			tests.ShortArraysStruct{}, "A", // Same name!
		)
	})
}
