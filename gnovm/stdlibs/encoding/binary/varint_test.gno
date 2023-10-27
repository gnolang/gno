package binary

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"testing"
)

func testConstant(t *testing.T, w uint, max int) {
	buf := make([]byte, MaxVarintLen64)
	n := PutUvarint(buf, 1<<w-1)
	if n != max {
		t.Errorf("MaxVarintLen%d = %d; want %d", w, max, n)
	}
}

func TestConstants(t *testing.T) {
	testConstant(t, 16, MaxVarintLen16)
	testConstant(t, 32, MaxVarintLen32)
	testConstant(t, 64, MaxVarintLen64)
}

// XXX SIGSEGV in test
// func testVarint(t *testing.T, x int64) {
// 	buf := make([]byte, MaxVarintLen64)
// 	n := PutVarint(buf, x)
// 	y, m := Varint(buf[0:n])
// 	if x != y {
// 		t.Errorf("Varint(%d): got %d", x, y)
// 	}
// 	if n != m {
// 		t.Errorf("Varint(%d): got n = %d; want %d", x, m, n)
// 	}

// 	buf2 := []byte("prefix")
// 	buf2 = AppendVarint(buf2, x)
// 	if string(buf2) != "prefix"+string(buf[:n]) {
// 		t.Errorf("AppendVarint(%d): got %q, want %q", x, buf2, "prefix"+string(buf[:n]))
// 	}

// 	y, err := ReadVarint(bytes.NewReader(buf))
// 	if err != nil {
// 		t.Errorf("ReadVarint(%d): %s", x, err)
// 	}
// 	if x != y {
// 		t.Errorf("ReadVarint(%d): got %d", x, y)
// 	}
// }

// func testUvarint(t *testing.T, x uint64) {
// 	buf := make([]byte, MaxVarintLen64)
// 	n := PutUvarint(buf, x)
// 	y, m := Uvarint(buf[0:n])
// 	if x != y {
// 		t.Errorf("Uvarint(%d): got %d", x, y)
// 	}
// 	if n != m {
// 		t.Errorf("Uvarint(%d): got n = %d; want %d", x, m, n)
// 	}

// 	buf2 := []byte("prefix")
// 	buf2 = AppendUvarint(buf2, x)
// 	if string(buf2) != "prefix"+string(buf[:n]) {
// 		t.Errorf("AppendUvarint(%d): got %q, want %q", x, buf2, "prefix"+string(buf[:n]))
// 	}

// 	y, err := ReadUvarint(bytes.NewReader(buf))
// 	if err != nil {
// 		t.Errorf("ReadUvarint(%d): %s", x, err)
// 	}
// 	if x != y {
// 		t.Errorf("ReadUvarint(%d): got %d", x, y)
// 	}
// }

// var tests = []int64{
// 	-1 << 63,
// 	-1<<63 + 1,
// 	-1,
// 	0,
// 	1,
// 	2,
// 	10,
// 	20,
// 	63,
// 	64,
// 	65,
// 	127,
// 	128,
// 	129,
// 	255,
// 	256,
// 	257,
// 	1<<63 - 1,
// }

// ref: https://pkg.go.dev/encoding/binary@go1.21.3#example-PutVarint
func TestPutVarint(t *testing.T) {
	buf := make([]byte, MaxVarintLen64)

	// testCases := []int64{
	// 	-1 << 63,
	// 	-1<<63 + 1,
	// 	-2,
	// 	-1,
	// 	0,
	// 	1,
	// 	2,
	// 	63,
	// 	64,
	// 	1<<63 - 1,
	// }
	// for _, tc := range testCases {
	// 	n := PutUvarint(buf, uint64(tc))
	// 	fmt.Printf("%x\n", buf[:n])
	// }
	// fmt.Printf("\n")

	testCases := []struct {
		input   int64
		expected string
	}{
		{-1<<63, "ffffffffffffffffff01"},
		{-1<<63+1, "fdffffffffffffffff01"},
		{-65, "8101"},
		{-64, "7f"},
		{-2, "03"},
		{-1, "01"},
		{0, "00"},
		{1, "02"},
		{2, "04"},
		{63, "7e"},
		{64, "8001"},
		{1<<63-1, "feffffffffffffffff01"},
	}

	for _, tc := range testCases {
		n := PutVarint(buf, tc.input)
		actual := ""
		for j := 0; j < n; j++ {
			actual += fmt.Sprintf("%02x", buf[j])
		}

		if actual != tc.expected {
			t.Errorf("PutVarint(%d): expected %s, got %s", tc.input, tc.expected, actual)
		}
	}
}

// ref: https://pkg.go.dev/encoding/binary@go1.21.3#example-PutUvarint
func TestPutUvarint(t *testing.T) {
	buf := make([]byte, MaxVarintLen64)

	testCases := []struct {
		input    uint64
		expected string
	}{
		{1, "01"},
		{2, "02"},
		{127, "7f"},
		{128, "8001"},
		{255, "ff01"},
		{256, "8002"},
		{1<<63 - 1, "ffffffffffffffff7f"},
	}

	for _, tc := range testCases {
		n := PutUvarint(buf, tc.input)
		actual := ""
		for j := 0; j < n; j++ {
			actual += fmt.Sprintf("%02x", buf[j])
		}

		if actual != tc.expected {
			t.Errorf("PutUvarint(%d): expected %s, got %s", tc.input, tc.expected, actual)
		}
	}
}

// ref: https://pkg.go.dev/encoding/binary@go1.21.3#example-Uvarint
func TestUvarint(t *testing.T) {
	testCases := []struct {
		input   []byte
		results uint64
	}{
		{[]byte{0x01}, 1},
		{[]byte{0x02}, 2},
		{[]byte{0x7f}, 127},
		{[]byte{0x80, 0x01}, 128},
		{[]byte{0xff, 0x01}, 255},
		{[]byte{0x80, 0x02}, 256},
	}

	for _, tc := range testCases {
		x, n := Uvarint(tc.input)
		if n != len(tc.input) {
			t.Errorf("Uvarint did not consume all bytes")
		}

		if x != tc.results {
			t.Errorf("Uvarint(%v): got x = %d, want %d", tc.input, x, tc.results)
		}
	}
}

func TestBufferTooSmall(t *testing.T) {
	buf := []byte{0x80, 0x80, 0x80, 0x80}
	for i := 0; i <= len(buf); i++ {
		buf := buf[0:i]
		x, n := Uvarint(buf)
		if x != 0 || n != 0 {
			t.Errorf("Uvarint(%v): got x = %d, n = %d", buf, x, n)
		}

		x, err := ReadUvarint(bytes.NewReader(buf))
		wantErr := io.EOF
		if i > 0 {
			wantErr = io.ErrUnexpectedEOF
		}
		if x != 0 || err != wantErr {
			t.Errorf("ReadUvarint(%v): got x = %d, err = %s", buf, x, err)
		}
	}
}

// XXX panic: reflect: reflect.Value.SetString using value obtained using unexported field
//
// Ensure that we catch overflows of bytes going past MaxVarintLen64.
// See issue https://golang.org/issues/41185
// func TestBufferTooBigWithOverflow(t *testing.T) {
// 	tests := []struct {
// 		in        []byte
// 		name      string
// 		wantN     int
// 		wantValue uint64
// 	}{
// 		{
// 			name: "invalid: 1000 bytes",
// 			in: func() []byte {
// 				b := make([]byte, 1000)
// 				for i := range b {
// 					b[i] = 0xff
// 				}
// 				b[999] = 0
// 				return b
// 			}(),
// 			wantN:     -11,
// 			wantValue: 0,
// 		},
// 		{
// 			name:      "valid: math.MaxUint64-40",
// 			in:        []byte{0xd7, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
// 			wantValue: math.MaxUint64 - 40,
// 			wantN:     10,
// 		},
// 		{
// 			name:      "invalid: with more than MaxVarintLen64 bytes",
// 			in:        []byte{0xd7, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
// 			wantN:     -11,
// 			wantValue: 0,
// 		},
// 		{
// 			name:      "invalid: 10th byte",
// 			in:        []byte{0xd7, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
// 			wantN:     -10,
// 			wantValue: 0,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			x, n := Uvarint(tt.in)
// 			if x != tt.wantValue || n != tt.wantN {
// 				t.Errorf("Uvarint(%v): got x = %d, n = %d; want %d, %d", tt.in, x, n, tt.wantValue, tt.wantN)
// 			}

// 			r := bytes.NewReader(tt.in)
// 			len := r.Len()
// 			x, err := ReadUvarint(r)
// 			if x != tt.wantValue || err != errOverflow {
// 				t.Errorf("ReadUvarint(%v): got x = %d, err = %s; want %d, %s", tt.in, x, err, tt.wantValue, errOverflow)
// 			}
// 			if read := len - r.Len(); read > MaxVarintLen64 {
// 				t.Errorf("ReadUvarint(%v): read more than MaxVarintLen64 bytes, got %d", tt.in, read)
// 			}
// 		})
// 	}
// }

func testOverflow(t *testing.T, buf []byte, x0 uint64, n0 int, err0 error) {
	x, n := Uvarint(buf)
	if x != 0 || n != n0 {
		t.Errorf("Uvarint(% X): got x = %d, n = %d; want 0, %d", buf, x, n, n0)
	}

	r := bytes.NewReader(buf)
	len := r.Len()
	x, err := ReadUvarint(r)
	if x != x0 || err != err0 {
		t.Errorf("ReadUvarint(%v): got x = %d, err = %s; want %d, %s", buf, x, err, x0, err0)
	}
	if read := len - r.Len(); read > MaxVarintLen64 {
		t.Errorf("ReadUvarint(%v): read more than MaxVarintLen64 bytes, got %d", buf, read)
	}
}

func TestOverflow(t *testing.T) {
	testOverflow(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x2}, 0, -10, errOverflow)
	testOverflow(t, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x1, 0, 0}, 0, -11, errOverflow)
	testOverflow(t, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 1<<64-1, -11, errOverflow) // 11 bytes, should overflow
}

func TestNonCanonicalZero(t *testing.T) {
	buf := []byte{0x80, 0x80, 0x80, 0}
	x, n := Uvarint(buf)
	if x != 0 || n != 4 {
		t.Errorf("Uvarint(%v): got x = %d, n = %d; want 0, 4", buf, x, n)
	}
}

func BenchmarkPutUvarint32(b *testing.B) {
	buf := make([]byte, MaxVarintLen32)
	b.SetBytes(4)
	for i := 0; i < b.N; i++ {
		for j := uint(0); j < MaxVarintLen32; j++ {
			PutUvarint(buf, 1<<(j*7))
		}
	}
}

func BenchmarkPutUvarint64(b *testing.B) {
	buf := make([]byte, MaxVarintLen64)
	b.SetBytes(8)
	for i := 0; i < b.N; i++ {
		for j := uint(0); j < MaxVarintLen64; j++ {
			PutUvarint(buf, 1<<(j*7))
		}
	}
}
