package binary_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
)

func TestLittleEndianUint16(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07}
	x1 := binary.LittleEndian.Uint16(b[0:])
	x2 := binary.LittleEndian.Uint16(b[2:])

	if x1 != 0x03e8 || x2 != 0x07d0 {
		t.Fatal("littleEndian.Uint16 failed")
	}
}

func TestLittleEndianPutUint16(t *testing.T) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint16(b[0:], 0x03e8)
	binary.LittleEndian.PutUint16(b[2:], 0x07d0)

	if !bytes.Equal(b, []byte{0xe8, 0x03, 0xd0, 0x07}) {
		t.Fatal("littleEndian.PutUint16 failed")
	}
}

func TestLittleEndianAppendUint16(t *testing.T) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint16(b[0:], 0x03e8)
	binary.LittleEndian.PutUint16(b[2:], 0x07d0)

	v := uint16(0x0bb8)
	b = binary.LittleEndian.AppendUint16(b, v)

	if !bytes.Equal(b, []byte{0xe8, 0x03, 0xd0, 0x07, 0xb8, 0x0b}) {
		t.Fatal("littleEndian.AppendUint16 failed")
	}
}

func TestLittleEndianUint32(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07}
	x := binary.LittleEndian.Uint32(b[0:])
	if x != 0x07d003e8 {
		t.Fatal("littleEndian.Uint32 failed")
	}
}

func TestLittleEndianPutUint32(t *testing.T) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b[0:], 0x07d003e8)

	if !bytes.Equal(b, []byte{0xe8, 0x03, 0xd0, 0x07}) {
		t.Fatal("littleEndian.PutUint32 failed")
	}
}

func TestLittleEndianAppendUint32(t *testing.T) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b[0:], 0x07d003e8)

	v := uint32(0x0bb807d0)
	b = binary.LittleEndian.AppendUint32(b, v)

	if !bytes.Equal(b, []byte{0xe8, 0x03, 0xd0, 0x07, 0xd0, 0x07, 0xb8, 0x0b}) {
		t.Fatal("littleEndian.AppendUint32 failed")
	}
}

func TestLittleEndianUint64(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07, 0xe9, 0xf5, 0x14, 0xfe}
	x := binary.LittleEndian.Uint64(b[0:])
	if x != 0xfe14f5e907d003e8 {
		t.Fatal("littleEndian.Uint64 failed")
	}
}

func TestLittleEndianPutUint64(t *testing.T) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b[0:], 0xfe14f5e907d003e8)

	if !bytes.Equal(b, []byte{0xe8, 0x03, 0xd0, 0x07, 0xe9, 0xf5, 0x14, 0xfe}) {
		t.Fatal("littleEndian.PutUint64 failed")
	}
}

func TestLittleEndianAppendUint64(t *testing.T) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b[0:], 0xfe14f5e907d003e8)

	v := uint64(0x0bb807d0d007e9f5)
	b = binary.LittleEndian.AppendUint64(b, v)

	if !bytes.Equal(b, []byte{0xe8, 0x03, 0xd0, 0x07, 0xe9, 0xf5, 0x14, 0xfe, 0xf5, 0xe9, 0x07, 0xd0, 0xd0, 0x07, 0xb8, 0x0b}) {
		t.Fatal("littleEndian.AppendUint64 failed")
	}
}

func TestBigEndianUint16(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07}
	x1 := binary.BigEndian.Uint16(b[0:])
	x2 := binary.BigEndian.Uint16(b[2:])

	if x1 != 0xe803 || x2 != 0xd007 {
		t.Fatal("bigEndian.Uint16 failed")
	}
}

func TestBigEndianPutUint16(t *testing.T) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[0:], 0x03e8)
	binary.BigEndian.PutUint16(b[2:], 0x07d0)

	if !bytes.Equal(b, []byte{0x03, 0xe8, 0x07, 0xd0}) {
		t.Fatal("bigEndian.PutUint16 failed")
	}
}

func TestBigEndianAppendUint16(t *testing.T) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[0:], 0x03e8)
	binary.BigEndian.PutUint16(b[2:], 0x07d0)

	v := uint16(0x0bb8)
	b = binary.BigEndian.AppendUint16(b, v)

	if !bytes.Equal(b, []byte{0x03, 0xe8, 0x07, 0xd0, 0x0b, 0xb8}) {
		t.Fatal("BigEndian.AppendUint16 failed")
	}
}

func TestBigEndianUint32(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07}
	x := binary.BigEndian.Uint32(b[0:])
	if x != 0xe803d007 {
		t.Fatal("bigEndian.Uint32 failed")
	}
}

func TestBigEndianPutUint32(t *testing.T) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[0:], 0x07d003e8)

	if !bytes.Equal(b, []byte{0x07, 0xd0, 0x03, 0xe8}) {
		t.Fatal("bigEndian.PutUint32 failed")
	}
}

func TestBigEndianAppendUint32(t *testing.T) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[0:], 0x07d003e8)

	v := uint32(0x0bb807d0)
	b = binary.BigEndian.AppendUint32(b, v)

	if !bytes.Equal(b, []byte{0x07, 0xd0, 0x03, 0xe8, 0x0b, 0xb8, 0x07, 0xd0}) {
		t.Fatal("bigEndian.AppendUint32 failed")
	}
}

func TestBigEndianUint64(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07, 0xe9, 0xf5, 0x14, 0xfe}
	x := binary.BigEndian.Uint64(b[0:])
	if x != 0xe803d007e9f514fe {
		t.Fatal("bigEndian.Uint64 failed")
	}
}

func TestBigEndianPutUint64(t *testing.T) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b[0:], 0xfe14f5e907d003e8)

	if !bytes.Equal(b, []byte{0xfe, 0x14, 0xf5, 0xe9, 0x07, 0xd0, 0x03, 0xe8}) {
		t.Fatal("bigEndian.PutUint64 failed")
	}
}

func TestbigEndianAppendUint64(t *testing.T) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b[0:], 0xfe14f5e907d003e8)

	v := uint64(0x0bb807d0d007e9f5)
	b = binary.BigEndian.AppendUint64(b, v)

	if !bytes.Equal(b, []byte{0xfe, 0x14, 0xf5, 0xe9, 0x07, 0xd0, 0x03, 0xe8, 0x0b, 0xb8, 0x07, 0xd0, 0xd0, 0x07, 0xe9, 0xf5}) {
		t.Fatal("bigEndian.AppendUint64 failed")
	}
}

// -----------------------------------------------------------------------------
// Tests ported from upstream go1.25.9 binary_test.go and example_test.go.
// Examples and tests that depend on binary.Read/Write/Encode/Decode/Append/Size
// are omitted: Gno's reflect package can't yet drive those generic codecs.
// -----------------------------------------------------------------------------

func testUint64SmallSliceLengthPanics() (panicked bool) {
	defer func() {
		panicked = recover() != nil
	}()
	b := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	binary.LittleEndian.Uint64(b[:4])
	return false
}

func testPutUint64SmallSliceLengthPanics() (panicked bool) {
	defer func() {
		panicked = recover() != nil
	}()
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:4], 0x0102030405060708)
	return false
}

func TestByteOrder(t *testing.T) {
	type byteOrder interface {
		binary.ByteOrder
		binary.AppendByteOrder
	}
	buf := make([]byte, 8)
	for _, order := range []byteOrder{binary.LittleEndian, binary.BigEndian} {
		const offset = 3
		for _, value := range []uint64{
			0x0000000000000000,
			0x0123456789abcdef,
			0xfedcba9876543210,
			0xffffffffffffffff,
			0xaaaaaaaaaaaaaaaa,
			math.Float64bits(math.Pi),
			math.Float64bits(math.E),
		} {
			want16 := uint16(value)
			order.PutUint16(buf[:2], want16)
			if got := order.Uint16(buf[:2]); got != want16 {
				t.Errorf("PutUint16: Uint16 = %v, want %v", got, want16)
			}
			buf = order.AppendUint16(buf[:offset], want16)
			if got := order.Uint16(buf[offset:]); got != want16 {
				t.Errorf("AppendUint16: Uint16 = %v, want %v", got, want16)
			}
			if len(buf) != offset+2 {
				t.Errorf("AppendUint16: len(buf) = %d, want %d", len(buf), offset+2)
			}

			want32 := uint32(value)
			order.PutUint32(buf[:4], want32)
			if got := order.Uint32(buf[:4]); got != want32 {
				t.Errorf("PutUint32: Uint32 = %v, want %v", got, want32)
			}
			buf = order.AppendUint32(buf[:offset], want32)
			if got := order.Uint32(buf[offset:]); got != want32 {
				t.Errorf("AppendUint32: Uint32 = %v, want %v", got, want32)
			}
			if len(buf) != offset+4 {
				t.Errorf("AppendUint32: len(buf) = %d, want %d", len(buf), offset+4)
			}

			want64 := uint64(value)
			order.PutUint64(buf[:8], want64)
			if got := order.Uint64(buf[:8]); got != want64 {
				t.Errorf("PutUint64: Uint64 = %v, want %v", got, want64)
			}
			buf = order.AppendUint64(buf[:offset], want64)
			if got := order.Uint64(buf[offset:]); got != want64 {
				t.Errorf("AppendUint64: Uint64 = %v, want %v", got, want64)
			}
			if len(buf) != offset+8 {
				t.Errorf("AppendUint64: len(buf) = %d, want %d", len(buf), offset+8)
			}
		}
	}
}

func TestEarlyBoundsChecks(t *testing.T) {
	if testUint64SmallSliceLengthPanics() != true {
		t.Errorf("binary.LittleEndian.Uint64 expected to panic for small slices, but didn't")
	}
	if testPutUint64SmallSliceLengthPanics() != true {
		t.Errorf("binary.LittleEndian.PutUint64 expected to panic for small slices, but didn't")
	}
}

// --- Examples (from upstream example_test.go) ---

func ExampleByteOrder_put() {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint16(b[0:], 0x03e8)
	binary.LittleEndian.PutUint16(b[2:], 0x07d0)
	fmt.Printf("% x\n", b)
	// Output:
	// e8 03 d0 07
}

func ExampleByteOrder_get() {
	b := []byte{0xe8, 0x03, 0xd0, 0x07}
	x1 := binary.LittleEndian.Uint16(b[0:])
	x2 := binary.LittleEndian.Uint16(b[2:])
	fmt.Printf("%#04x %#04x\n", x1, x2)
	// Output:
	// 0x03e8 0x07d0
}

func ExamplePutUvarint() {
	buf := make([]byte, binary.MaxVarintLen64)

	for _, x := range []uint64{1, 2, 127, 128, 255, 256} {
		n := binary.PutUvarint(buf, x)
		fmt.Printf("%x\n", buf[:n])
	}
	// Output:
	// 01
	// 02
	// 7f
	// 8001
	// ff01
	// 8002
}

func ExamplePutVarint() {
	buf := make([]byte, binary.MaxVarintLen64)

	for _, x := range []int64{-65, -64, -2, -1, 0, 1, 2, 63, 64} {
		n := binary.PutVarint(buf, x)
		fmt.Printf("%x\n", buf[:n])
	}
	// Output:
	// 8101
	// 7f
	// 03
	// 01
	// 00
	// 02
	// 04
	// 7e
	// 8001
}

func ExampleUvarint() {
	inputs := [][]byte{
		{0x01},
		{0x02},
		{0x7f},
		{0x80, 0x01},
		{0xff, 0x01},
		{0x80, 0x02},
	}
	for _, b := range inputs {
		x, n := binary.Uvarint(b)
		if n != len(b) {
			fmt.Println("Uvarint did not consume all of in")
		}
		fmt.Println(x)
	}
	// Output:
	// 1
	// 2
	// 127
	// 128
	// 255
	// 256
}

func ExampleVarint() {
	inputs := [][]byte{
		{0x81, 0x01},
		{0x7f},
		{0x03},
		{0x01},
		{0x00},
		{0x02},
		{0x04},
		{0x7e},
		{0x80, 0x01},
	}
	for _, b := range inputs {
		x, n := binary.Varint(b)
		if n != len(b) {
			fmt.Println("Varint did not consume all of in")
		}
		fmt.Println(x)
	}
	// Output:
	// -65
	// -64
	// -2
	// -1
	// 0
	// 1
	// 2
	// 63
	// 64
}
