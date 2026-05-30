// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bits_test

import (
	"fmt"
	"math/bits"
	"testing"
)

func TestLeadingZeros(t *testing.T) {
	for i := 0; i < 256; i++ {
		nlz := tab[i].nlz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			if x <= 1<<8-1 {
				got := bits.LeadingZeros8(uint8(x))
				want := nlz - k + (8 - 8)
				if x == 0 {
					want = 8
				}
				if got != want {
					t.Fatalf("LeadingZeros8(%#02x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := bits.LeadingZeros16(uint16(x))
				want := nlz - k + (16 - 8)
				if x == 0 {
					want = 16
				}
				if got != want {
					t.Fatalf("LeadingZeros16(%#04x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := bits.LeadingZeros32(uint32(x))
				want := nlz - k + (32 - 8)
				if x == 0 {
					want = 32
				}
				if got != want {
					t.Fatalf("LeadingZeros32(%#08x) == %d; want %d", x, got, want)
				}
				if bits.UintSize == 32 {
					got = bits.LeadingZeros(uint(x))
					if got != want {
						t.Fatalf("LeadingZeros(%#08x) == %d; want %d", x, got, want)
					}
				}
			}

			if x <= 1<<64-1 {
				got := bits.LeadingZeros64(uint64(x))
				want := nlz - k + (64 - 8)
				if x == 0 {
					want = 64
				}
				if got != want {
					t.Fatalf("LeadingZeros64(%#016x) == %d; want %d", x, got, want)
				}
				if bits.UintSize == 64 {
					got = bits.LeadingZeros(uint(x))
					if got != want {
						t.Fatalf("LeadingZeros(%#016x) == %d; want %d", x, got, want)
					}
				}
			}
		}
	}
}

// Exported (global) variable serving as input for some
// of the benchmarks to ensure side-effect free calls
// are not optimized away.
var Input uint64 = bits.DeBruijn64

// Exported (global) variable to store function results
// during benchmarking to ensure side-effect free calls
// are not optimized away.
var Output int

func BenchmarkLeadingZeros(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.LeadingZeros(uint(Input) >> (uint(i) % bits.UintSize))
	}
	Output = s
}

func BenchmarkLeadingZeros8(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.LeadingZeros8(uint8(Input) >> (uint(i) % 8))
	}
	Output = s
}

func BenchmarkLeadingZeros16(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.LeadingZeros16(uint16(Input) >> (uint(i) % 16))
	}
	Output = s
}

func BenchmarkLeadingZeros32(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.LeadingZeros32(uint32(Input) >> (uint(i) % 32))
	}
	Output = s
}

func BenchmarkLeadingZeros64(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.LeadingZeros64(uint64(Input) >> (uint(i) % 64))
	}
	Output = s
}

func TestTrailingZeros(t *testing.T) {
	for i := 0; i < 256; i++ {
		ntz := tab[i].ntz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			want := ntz + k
			if x <= 1<<8-1 {
				got := bits.TrailingZeros8(uint8(x))
				if x == 0 {
					want = 8
				}
				if got != want {
					t.Fatalf("TrailingZeros8(%#02x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := bits.TrailingZeros16(uint16(x))
				if x == 0 {
					want = 16
				}
				if got != want {
					t.Fatalf("TrailingZeros16(%#04x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := bits.TrailingZeros32(uint32(x))
				if x == 0 {
					want = 32
				}
				if got != want {
					t.Fatalf("TrailingZeros32(%#08x) == %d; want %d", x, got, want)
				}
				if bits.UintSize == 32 {
					got = bits.TrailingZeros(uint(x))
					if got != want {
						t.Fatalf("TrailingZeros(%#08x) == %d; want %d", x, got, want)
					}
				}
			}

			if x <= 1<<64-1 {
				got := bits.TrailingZeros64(uint64(x))
				if x == 0 {
					want = 64
				}
				if got != want {
					t.Fatalf("TrailingZeros64(%#016x) == %d; want %d", x, got, want)
				}
				if bits.UintSize == 64 {
					got = bits.TrailingZeros(uint(x))
					if got != want {
						t.Fatalf("TrailingZeros(%#016x) == %d; want %d", x, got, want)
					}
				}
			}
		}
	}
}

func BenchmarkTrailingZeros(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.TrailingZeros(uint(Input) << (uint(i) % bits.UintSize))
	}
	Output = s
}

func BenchmarkTrailingZeros8(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.TrailingZeros8(uint8(Input) << (uint(i) % 8))
	}
	Output = s
}

func BenchmarkTrailingZeros16(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.TrailingZeros16(uint16(Input) << (uint(i) % 16))
	}
	Output = s
}

func BenchmarkTrailingZeros32(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.TrailingZeros32(uint32(Input) << (uint(i) % 32))
	}
	Output = s
}

func BenchmarkTrailingZeros64(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.TrailingZeros64(uint64(Input) << (uint(i) % 64))
	}
	Output = s
}

func TestOnesCount(t *testing.T) {
	var x uint64
	for i := 0; i <= 64; i++ {
		testOnesCount(t, x, i)
		x = x<<1 | 1
	}

	for i := 64; i >= 0; i-- {
		testOnesCount(t, x, i)
		x = x << 1
	}

	for i := 0; i < 256; i++ {
		for k := 0; k < 64-8; k++ {
			testOnesCount(t, uint64(i)<<uint(k), tab[i].pop)
		}
	}
}

func testOnesCount(t *testing.T, x uint64, want int) {
	if x <= 1<<8-1 {
		got := bits.OnesCount8(uint8(x))
		if got != want {
			t.Fatalf("OnesCount8(%#02x) == %d; want %d", uint8(x), got, want)
		}
	}

	if x <= 1<<16-1 {
		got := bits.OnesCount16(uint16(x))
		if got != want {
			t.Fatalf("OnesCount16(%#04x) == %d; want %d", uint16(x), got, want)
		}
	}

	if x <= 1<<32-1 {
		got := bits.OnesCount32(uint32(x))
		if got != want {
			t.Fatalf("OnesCount32(%#08x) == %d; want %d", uint32(x), got, want)
		}
		if bits.UintSize == 32 {
			got = bits.OnesCount(uint(x))
			if got != want {
				t.Fatalf("OnesCount(%#08x) == %d; want %d", uint32(x), got, want)
			}
		}
	}

	if x <= 1<<64-1 {
		got := bits.OnesCount64(uint64(x))
		if got != want {
			t.Fatalf("OnesCount64(%#016x) == %d; want %d", x, got, want)
		}
		if bits.UintSize == 64 {
			got = bits.OnesCount(uint(x))
			if got != want {
				t.Fatalf("OnesCount(%#016x) == %d; want %d", x, got, want)
			}
		}
	}
}

func BenchmarkOnesCount(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.OnesCount(uint(Input))
	}
	Output = s
}

func BenchmarkOnesCount8(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.OnesCount8(uint8(Input))
	}
	Output = s
}

func BenchmarkOnesCount16(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.OnesCount16(uint16(Input))
	}
	Output = s
}

func BenchmarkOnesCount32(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.OnesCount32(uint32(Input))
	}
	Output = s
}

func BenchmarkOnesCount64(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += bits.OnesCount64(uint64(Input))
	}
	Output = s
}

func TestRotateLeft(t *testing.T) {
	var m uint64 = bits.DeBruijn64

	for k := uint(0); k < 128; k++ {
		x8 := uint8(m)
		got8 := bits.RotateLeft8(x8, int(k))
		want8 := x8<<(k&0x7) | x8>>(8-k&0x7)
		if got8 != want8 {
			t.Fatalf("RotateLeft8(%#02x, %d) == %#02x; want %#02x", x8, k, got8, want8)
		}
		got8 = bits.RotateLeft8(want8, -int(k))
		if got8 != x8 {
			t.Fatalf("RotateLeft8(%#02x, -%d) == %#02x; want %#02x", want8, k, got8, x8)
		}

		x16 := uint16(m)
		got16 := bits.RotateLeft16(x16, int(k))
		want16 := x16<<(k&0xf) | x16>>(16-k&0xf)
		if got16 != want16 {
			t.Fatalf("RotateLeft16(%#04x, %d) == %#04x; want %#04x", x16, k, got16, want16)
		}
		got16 = bits.RotateLeft16(want16, -int(k))
		if got16 != x16 {
			t.Fatalf("RotateLeft16(%#04x, -%d) == %#04x; want %#04x", want16, k, got16, x16)
		}

		x32 := uint32(m)
		got32 := bits.RotateLeft32(x32, int(k))
		want32 := x32<<(k&0x1f) | x32>>(32-k&0x1f)
		if got32 != want32 {
			t.Fatalf("RotateLeft32(%#08x, %d) == %#08x; want %#08x", x32, k, got32, want32)
		}
		got32 = bits.RotateLeft32(want32, -int(k))
		if got32 != x32 {
			t.Fatalf("RotateLeft32(%#08x, -%d) == %#08x; want %#08x", want32, k, got32, x32)
		}
		if bits.UintSize == 32 {
			x := uint(m)
			got := bits.RotateLeft(x, int(k))
			want := x<<(k&0x1f) | x>>(32-k&0x1f)
			if got != want {
				t.Fatalf("RotateLeft(%#08x, %d) == %#08x; want %#08x", x, k, got, want)
			}
			got = bits.RotateLeft(want, -int(k))
			if got != x {
				t.Fatalf("RotateLeft(%#08x, -%d) == %#08x; want %#08x", want, k, got, x)
			}
		}

		x64 := uint64(m)
		got64 := bits.RotateLeft64(x64, int(k))
		want64 := x64<<(k&0x3f) | x64>>(64-k&0x3f)
		if got64 != want64 {
			t.Fatalf("RotateLeft64(%#016x, %d) == %#016x; want %#016x", x64, k, got64, want64)
		}
		got64 = bits.RotateLeft64(want64, -int(k))
		if got64 != x64 {
			t.Fatalf("RotateLeft64(%#016x, -%d) == %#016x; want %#016x", want64, k, got64, x64)
		}
		if bits.UintSize == 64 {
			x := uint(m)
			got := bits.RotateLeft(x, int(k))
			want := x<<(k&0x3f) | x>>(64-k&0x3f)
			if got != want {
				t.Fatalf("RotateLeft(%#016x, %d) == %#016x; want %#016x", x, k, got, want)
			}
			got = bits.RotateLeft(want, -int(k))
			if got != x {
				t.Fatalf("RotateLeft(%#08x, -%d) == %#08x; want %#08x", want, k, got, x)
			}
		}
	}
}

func BenchmarkRotateLeft(b *testing.B) {
	var s uint
	for i := 0; i < b.N; i++ {
		s += bits.RotateLeft(uint(Input), i)
	}
	Output = int(s)
}

func BenchmarkRotateLeft8(b *testing.B) {
	var s uint8
	for i := 0; i < b.N; i++ {
		s += bits.RotateLeft8(uint8(Input), i)
	}
	Output = int(s)
}

func BenchmarkRotateLeft16(b *testing.B) {
	var s uint16
	for i := 0; i < b.N; i++ {
		s += bits.RotateLeft16(uint16(Input), i)
	}
	Output = int(s)
}

func BenchmarkRotateLeft32(b *testing.B) {
	var s uint32
	for i := 0; i < b.N; i++ {
		s += bits.RotateLeft32(uint32(Input), i)
	}
	Output = int(s)
}

func BenchmarkRotateLeft64(b *testing.B) {
	var s uint64
	for i := 0; i < b.N; i++ {
		s += bits.RotateLeft64(uint64(Input), i)
	}
	Output = int(s)
}

func TestReverse(t *testing.T) {
	// test each bit
	for i := uint(0); i < 64; i++ {
		testReverse(t, uint64(1)<<i, uint64(1)<<(63-i))
	}

	// test a few patterns
	for _, test := range []struct {
		x, r uint64
	}{
		{0, 0},
		{0x1, 0x8 << 60},
		{0x2, 0x4 << 60},
		{0x3, 0xc << 60},
		{0x4, 0x2 << 60},
		{0x5, 0xa << 60},
		{0x6, 0x6 << 60},
		{0x7, 0xe << 60},
		{0x8, 0x1 << 60},
		{0x9, 0x9 << 60},
		{0xa, 0x5 << 60},
		{0xb, 0xd << 60},
		{0xc, 0x3 << 60},
		{0xd, 0xb << 60},
		{0xe, 0x7 << 60},
		{0xf, 0xf << 60},
		{0x5686487, 0xe12616a000000000},
		{0x0123456789abcdef, 0xf7b3d591e6a2c480},
	} {
		testReverse(t, test.x, test.r)
		testReverse(t, test.r, test.x)
	}
}

func testReverse(t *testing.T, x64, want64 uint64) {
	x8 := uint8(x64)
	got8 := bits.Reverse8(x8)
	want8 := uint8(want64 >> (64 - 8))
	if got8 != want8 {
		t.Fatalf("Reverse8(%#02x) == %#02x; want %#02x", x8, got8, want8)
	}

	x16 := uint16(x64)
	got16 := bits.Reverse16(x16)
	want16 := uint16(want64 >> (64 - 16))
	if got16 != want16 {
		t.Fatalf("Reverse16(%#04x) == %#04x; want %#04x", x16, got16, want16)
	}

	x32 := uint32(x64)
	got32 := bits.Reverse32(x32)
	want32 := uint32(want64 >> (64 - 32))
	if got32 != want32 {
		t.Fatalf("Reverse32(%#08x) == %#08x; want %#08x", x32, got32, want32)
	}
	if bits.UintSize == 32 {
		x := uint(x32)
		got := bits.Reverse(x)
		want := uint(want32)
		if got != want {
			t.Fatalf("Reverse(%#08x) == %#08x; want %#08x", x, got, want)
		}
	}

	got64 := bits.Reverse64(x64)
	if got64 != want64 {
		t.Fatalf("Reverse64(%#016x) == %#016x; want %#016x", x64, got64, want64)
	}
	if bits.UintSize == 64 {
		x := uint(x64)
		got := bits.Reverse(x)
		want := uint(want64)
		if got != want {
			t.Fatalf("Reverse(%#08x) == %#016x; want %#016x", x, got, want)
		}
	}
}

func BenchmarkReverse(b *testing.B) {
	var s uint
	for i := 0; i < b.N; i++ {
		s += bits.Reverse(uint(i))
	}
	Output = int(s)
}

func BenchmarkReverse8(b *testing.B) {
	var s uint8
	for i := 0; i < b.N; i++ {
		s += bits.Reverse8(uint8(i))
	}
	Output = int(s)
}

func BenchmarkReverse16(b *testing.B) {
	var s uint16
	for i := 0; i < b.N; i++ {
		s += bits.Reverse16(uint16(i))
	}
	Output = int(s)
}

func BenchmarkReverse32(b *testing.B) {
	var s uint32
	for i := 0; i < b.N; i++ {
		s += bits.Reverse32(uint32(i))
	}
	Output = int(s)
}

func BenchmarkReverse64(b *testing.B) {
	var s uint64
	for i := 0; i < b.N; i++ {
		s += bits.Reverse64(uint64(i))
	}
	Output = int(s)
}

func TestReverseBytes(t *testing.T) {
	for _, test := range []struct {
		x, r uint64
	}{
		{0, 0},
		{0x01, 0x01 << 56},
		{0x0123, 0x2301 << 48},
		{0x012345, 0x452301 << 40},
		{0x01234567, 0x67452301 << 32},
		{0x0123456789, 0x8967452301 << 24},
		{0x0123456789ab, 0xab8967452301 << 16},
		{0x0123456789abcd, 0xcdab8967452301 << 8},
		{0x0123456789abcdef, 0xefcdab8967452301 << 0},
	} {
		testReverseBytes(t, test.x, test.r)
		testReverseBytes(t, test.r, test.x)
	}
}

func testReverseBytes(t *testing.T, x64, want64 uint64) {
	x16 := uint16(x64)
	got16 := bits.ReverseBytes16(x16)
	want16 := uint16(want64 >> (64 - 16))
	if got16 != want16 {
		t.Fatalf("ReverseBytes16(%#04x) == %#04x; want %#04x", x16, got16, want16)
	}

	x32 := uint32(x64)
	got32 := bits.ReverseBytes32(x32)
	want32 := uint32(want64 >> (64 - 32))
	if got32 != want32 {
		t.Fatalf("ReverseBytes32(%#08x) == %#08x; want %#08x", x32, got32, want32)
	}
	if bits.UintSize == 32 {
		x := uint(x32)
		got := bits.ReverseBytes(x)
		want := uint(want32)
		if got != want {
			t.Fatalf("ReverseBytes(%#08x) == %#08x; want %#08x", x, got, want)
		}
	}

	got64 := bits.ReverseBytes64(x64)
	if got64 != want64 {
		t.Fatalf("ReverseBytes64(%#016x) == %#016x; want %#016x", x64, got64, want64)
	}
	if bits.UintSize == 64 {
		x := uint(x64)
		got := bits.ReverseBytes(x)
		want := uint(want64)
		if got != want {
			t.Fatalf("ReverseBytes(%#016x) == %#016x; want %#016x", x, got, want)
		}
	}
}

func BenchmarkReverseBytes(b *testing.B) {
	var s uint
	for i := 0; i < b.N; i++ {
		s += bits.ReverseBytes(uint(i))
	}
	Output = int(s)
}

func BenchmarkReverseBytes16(b *testing.B) {
	var s uint16
	for i := 0; i < b.N; i++ {
		s += bits.ReverseBytes16(uint16(i))
	}
	Output = int(s)
}

func BenchmarkReverseBytes32(b *testing.B) {
	var s uint32
	for i := 0; i < b.N; i++ {
		s += bits.ReverseBytes32(uint32(i))
	}
	Output = int(s)
}

func BenchmarkReverseBytes64(b *testing.B) {
	var s uint64
	for i := 0; i < b.N; i++ {
		s += bits.ReverseBytes64(uint64(i))
	}
	Output = int(s)
}

func TestLen(t *testing.T) {
	for i := 0; i < 256; i++ {
		length := 8 - tab[i].nlz
		for k := 0; k < 64-8; k++ {
			x := uint64(i) << uint(k)
			want := 0
			if x != 0 {
				want = length + k
			}
			if x <= 1<<8-1 {
				got := bits.Len8(uint8(x))
				if got != want {
					t.Fatalf("Len8(%#02x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<16-1 {
				got := bits.Len16(uint16(x))
				if got != want {
					t.Fatalf("Len16(%#04x) == %d; want %d", x, got, want)
				}
			}

			if x <= 1<<32-1 {
				got := bits.Len32(uint32(x))
				if got != want {
					t.Fatalf("Len32(%#08x) == %d; want %d", x, got, want)
				}
				if bits.UintSize == 32 {
					got := bits.Len(uint(x))
					if got != want {
						t.Fatalf("Len(%#08x) == %d; want %d", x, got, want)
					}
				}
			}

			if x <= 1<<64-1 {
				got := bits.Len64(uint64(x))
				if got != want {
					t.Fatalf("Len64(%#016x) == %d; want %d", x, got, want)
				}
				if bits.UintSize == 64 {
					got := bits.Len(uint(x))
					if got != want {
						t.Fatalf("Len(%#016x) == %d; want %d", x, got, want)
					}
				}
			}
		}
	}
}

const (
	_M   = 1<<bits.UintSize - 1
	_M32 = 1<<32 - 1
	_M64 = 1<<64 - 1
)

func TestAddSubUint(t *testing.T) {
	test := func(msg string, f func(x, y, c uint) (z, cout uint), x, y, c, z, cout uint) {
		z1, cout1 := f(x, y, c)
		if z1 != z || cout1 != cout {
			t.Errorf("%s: got z:cout = %#x:%#x; want %#x:%#x", msg, z1, cout1, z, cout)
		}
	}
	for _, a := range []struct{ x, y, c, z, cout uint }{
		{0, 0, 0, 0, 0},
		{0, 1, 0, 1, 0},
		{0, 0, 1, 1, 0},
		{0, 1, 1, 2, 0},
		{12345, 67890, 0, 80235, 0},
		{12345, 67890, 1, 80236, 0},
		{_M, 1, 0, 0, 1},
		{_M, 0, 1, 0, 1},
		{_M, 1, 1, 1, 1},
		{_M, _M, 0, _M - 1, 1},
		{_M, _M, 1, _M, 1},
	} {
		test("Add", bits.Add, a.x, a.y, a.c, a.z, a.cout)
		test("Add symmetric", bits.Add, a.y, a.x, a.c, a.z, a.cout)
		test("Sub", bits.Sub, a.z, a.x, a.c, a.y, a.cout)
		test("Sub symmetric", bits.Sub, a.z, a.y, a.c, a.x, a.cout)
		// The above code can't test intrinsic implementation, because the passed function is not called directly.
		// The following code uses a closure to test the intrinsic version in case the function is intrinsified.
		test("Add intrinsic", func(x, y, c uint) (uint, uint) { return bits.Add(x, y, c) }, a.x, a.y, a.c, a.z, a.cout)
		test("Add intrinsic symmetric", func(x, y, c uint) (uint, uint) { return bits.Add(x, y, c) }, a.y, a.x, a.c, a.z, a.cout)
		test("Sub intrinsic", func(x, y, c uint) (uint, uint) { return bits.Sub(x, y, c) }, a.z, a.x, a.c, a.y, a.cout)
		test("Sub intrinsic symmetric", func(x, y, c uint) (uint, uint) { return bits.Sub(x, y, c) }, a.z, a.y, a.c, a.x, a.cout)

	}
}

func TestAddSubUint32(t *testing.T) {
	test := func(msg string, f func(x, y, c uint32) (z, cout uint32), x, y, c, z, cout uint32) {
		z1, cout1 := f(x, y, c)
		if z1 != z || cout1 != cout {
			t.Errorf("%s: got z:cout = %#x:%#x; want %#x:%#x", msg, z1, cout1, z, cout)
		}
	}
	for _, a := range []struct{ x, y, c, z, cout uint32 }{
		{0, 0, 0, 0, 0},
		{0, 1, 0, 1, 0},
		{0, 0, 1, 1, 0},
		{0, 1, 1, 2, 0},
		{12345, 67890, 0, 80235, 0},
		{12345, 67890, 1, 80236, 0},
		{_M32, 1, 0, 0, 1},
		{_M32, 0, 1, 0, 1},
		{_M32, 1, 1, 1, 1},
		{_M32, _M32, 0, _M32 - 1, 1},
		{_M32, _M32, 1, _M32, 1},
	} {
		test("Add32", bits.Add32, a.x, a.y, a.c, a.z, a.cout)
		test("Add32 symmetric", bits.Add32, a.y, a.x, a.c, a.z, a.cout)
		test("Sub32", bits.Sub32, a.z, a.x, a.c, a.y, a.cout)
		test("Sub32 symmetric", bits.Sub32, a.z, a.y, a.c, a.x, a.cout)
	}
}

func TestAddSubUint64(t *testing.T) {
	test := func(msg string, f func(x, y, c uint64) (z, cout uint64), x, y, c, z, cout uint64) {
		z1, cout1 := f(x, y, c)
		if z1 != z || cout1 != cout {
			t.Errorf("%s: got z:cout = %#x:%#x; want %#x:%#x", msg, z1, cout1, z, cout)
		}
	}
	for _, a := range []struct{ x, y, c, z, cout uint64 }{
		{0, 0, 0, 0, 0},
		{0, 1, 0, 1, 0},
		{0, 0, 1, 1, 0},
		{0, 1, 1, 2, 0},
		{12345, 67890, 0, 80235, 0},
		{12345, 67890, 1, 80236, 0},
		{_M64, 1, 0, 0, 1},
		{_M64, 0, 1, 0, 1},
		{_M64, 1, 1, 1, 1},
		{_M64, _M64, 0, _M64 - 1, 1},
		{_M64, _M64, 1, _M64, 1},
	} {
		test("Add64", bits.Add64, a.x, a.y, a.c, a.z, a.cout)
		test("Add64 symmetric", bits.Add64, a.y, a.x, a.c, a.z, a.cout)
		test("Sub64", bits.Sub64, a.z, a.x, a.c, a.y, a.cout)
		test("Sub64 symmetric", bits.Sub64, a.z, a.y, a.c, a.x, a.cout)
		// The above code can't test intrinsic implementation, because the passed function is not called directly.
		// The following code uses a closure to test the intrinsic version in case the function is intrinsified.
		test("Add64 intrinsic", func(x, y, c uint64) (uint64, uint64) { return bits.Add64(x, y, c) }, a.x, a.y, a.c, a.z, a.cout)
		test("Add64 intrinsic symmetric", func(x, y, c uint64) (uint64, uint64) { return bits.Add64(x, y, c) }, a.y, a.x, a.c, a.z, a.cout)
		test("Sub64 intrinsic", func(x, y, c uint64) (uint64, uint64) { return bits.Sub64(x, y, c) }, a.z, a.x, a.c, a.y, a.cout)
		test("Sub64 intrinsic symmetric", func(x, y, c uint64) (uint64, uint64) { return bits.Sub64(x, y, c) }, a.z, a.y, a.c, a.x, a.cout)
	}
}

func TestAdd64OverflowPanic(t *testing.T) {
	// Test that 64-bit overflow panics fire correctly.
	// These are designed to improve coverage of compiler intrinsics.
	tests := []func(uint64, uint64) uint64{
		func(a, b uint64) uint64 {
			x, c := bits.Add64(a, b, 0)
			if c > 0 {
				panic("overflow")
			}
			return x
		},
		func(a, b uint64) uint64 {
			x, c := bits.Add64(a, b, 0)
			if c != 0 {
				panic("overflow")
			}
			return x
		},
		func(a, b uint64) uint64 {
			x, c := bits.Add64(a, b, 0)
			if c == 1 {
				panic("overflow")
			}
			return x
		},
		func(a, b uint64) uint64 {
			x, c := bits.Add64(a, b, 0)
			if c != 1 {
				return x
			}
			panic("overflow")
		},
		func(a, b uint64) uint64 {
			x, c := bits.Add64(a, b, 0)
			if c == 0 {
				return x
			}
			panic("overflow")
		},
	}
	for _, test := range tests {
		shouldPanic := func(f func()) {
			defer func() {
				if err := recover(); err == nil {
					t.Fatalf("expected panic")
				}
			}()
			f()
		}

		// overflow
		shouldPanic(func() { test(_M64, 1) })
		shouldPanic(func() { test(1, _M64) })
		shouldPanic(func() { test(_M64, _M64) })

		// no overflow
		test(_M64, 0)
		test(0, 0)
		test(1, 1)
	}
}

func TestSub64OverflowPanic(t *testing.T) {
	// Test that 64-bit overflow panics fire correctly.
	// These are designed to improve coverage of compiler intrinsics.
	tests := []func(uint64, uint64) uint64{
		func(a, b uint64) uint64 {
			x, c := bits.Sub64(a, b, 0)
			if c > 0 {
				panic("overflow")
			}
			return x
		},
		func(a, b uint64) uint64 {
			x, c := bits.Sub64(a, b, 0)
			if c != 0 {
				panic("overflow")
			}
			return x
		},
		func(a, b uint64) uint64 {
			x, c := bits.Sub64(a, b, 0)
			if c == 1 {
				panic("overflow")
			}
			return x
		},
		func(a, b uint64) uint64 {
			x, c := bits.Sub64(a, b, 0)
			if c != 1 {
				return x
			}
			panic("overflow")
		},
		func(a, b uint64) uint64 {
			x, c := bits.Sub64(a, b, 0)
			if c == 0 {
				return x
			}
			panic("overflow")
		},
	}
	for _, test := range tests {
		shouldPanic := func(f func()) {
			defer func() {
				if err := recover(); err == nil {
					t.Fatalf("expected panic")
				}
			}()
			f()
		}

		// overflow
		shouldPanic(func() { test(0, 1) })
		shouldPanic(func() { test(1, _M64) })
		shouldPanic(func() { test(_M64-1, _M64) })

		// no overflow
		test(_M64, 0)
		test(0, 0)
		test(1, 1)
	}
}

func TestMulDiv(t *testing.T) {
	testMul := func(msg string, f func(x, y uint) (hi, lo uint), x, y, hi, lo uint) {
		hi1, lo1 := f(x, y)
		if hi1 != hi || lo1 != lo {
			t.Errorf("%s: got hi:lo = %#x:%#x; want %#x:%#x", msg, hi1, lo1, hi, lo)
		}
	}
	testDiv := func(msg string, f func(hi, lo, y uint) (q, r uint), hi, lo, y, q, r uint) {
		q1, r1 := f(hi, lo, y)
		if q1 != q || r1 != r {
			t.Errorf("%s: got q:r = %#x:%#x; want %#x:%#x", msg, q1, r1, q, r)
		}
	}
	for _, a := range []struct {
		x, y      uint
		hi, lo, r uint
	}{
		{1 << (bits.UintSize - 1), 2, 1, 0, 1},
		{_M, _M, _M - 1, 1, 42},
	} {
		testMul("Mul", bits.Mul, a.x, a.y, a.hi, a.lo)
		testMul("Mul symmetric", bits.Mul, a.y, a.x, a.hi, a.lo)
		testDiv("Div", bits.Div, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div symmetric", bits.Div, a.hi, a.lo+a.r, a.x, a.y, a.r)
		// The above code can't test intrinsic implementation, because the passed function is not called directly.
		// The following code uses a closure to test the intrinsic version in case the function is intrinsified.
		testMul("Mul intrinsic", func(x, y uint) (uint, uint) { return bits.Mul(x, y) }, a.x, a.y, a.hi, a.lo)
		testMul("Mul intrinsic symmetric", func(x, y uint) (uint, uint) { return bits.Mul(x, y) }, a.y, a.x, a.hi, a.lo)
		testDiv("Div intrinsic", func(hi, lo, y uint) (uint, uint) { return bits.Div(hi, lo, y) }, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div intrinsic symmetric", func(hi, lo, y uint) (uint, uint) { return bits.Div(hi, lo, y) }, a.hi, a.lo+a.r, a.x, a.y, a.r)
	}
}

func TestMulDiv32(t *testing.T) {
	testMul := func(msg string, f func(x, y uint32) (hi, lo uint32), x, y, hi, lo uint32) {
		hi1, lo1 := f(x, y)
		if hi1 != hi || lo1 != lo {
			t.Errorf("%s: got hi:lo = %#x:%#x; want %#x:%#x", msg, hi1, lo1, hi, lo)
		}
	}
	testDiv := func(msg string, f func(hi, lo, y uint32) (q, r uint32), hi, lo, y, q, r uint32) {
		q1, r1 := f(hi, lo, y)
		if q1 != q || r1 != r {
			t.Errorf("%s: got q:r = %#x:%#x; want %#x:%#x", msg, q1, r1, q, r)
		}
	}
	for _, a := range []struct {
		x, y      uint32
		hi, lo, r uint32
	}{
		{1 << 31, 2, 1, 0, 1},
		{0xc47dfa8c, 50911, 0x98a4, 0x998587f4, 13},
		{_M32, _M32, _M32 - 1, 1, 42},
	} {
		testMul("Mul32", bits.Mul32, a.x, a.y, a.hi, a.lo)
		testMul("Mul32 symmetric", bits.Mul32, a.y, a.x, a.hi, a.lo)
		testDiv("Div32", bits.Div32, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div32 symmetric", bits.Div32, a.hi, a.lo+a.r, a.x, a.y, a.r)
	}
}

func TestMulDiv64(t *testing.T) {
	testMul := func(msg string, f func(x, y uint64) (hi, lo uint64), x, y, hi, lo uint64) {
		hi1, lo1 := f(x, y)
		if hi1 != hi || lo1 != lo {
			t.Errorf("%s: got hi:lo = %#x:%#x; want %#x:%#x", msg, hi1, lo1, hi, lo)
		}
	}
	testDiv := func(msg string, f func(hi, lo, y uint64) (q, r uint64), hi, lo, y, q, r uint64) {
		q1, r1 := f(hi, lo, y)
		if q1 != q || r1 != r {
			t.Errorf("%s: got q:r = %#x:%#x; want %#x:%#x", msg, q1, r1, q, r)
		}
	}
	for _, a := range []struct {
		x, y      uint64
		hi, lo, r uint64
	}{
		{1 << 63, 2, 1, 0, 1},
		{0x3626229738a3b9, 0xd8988a9f1cc4a61, 0x2dd0712657fe8, 0x9dd6a3364c358319, 13},
		{_M64, _M64, _M64 - 1, 1, 42},
	} {
		testMul("Mul64", bits.Mul64, a.x, a.y, a.hi, a.lo)
		testMul("Mul64 symmetric", bits.Mul64, a.y, a.x, a.hi, a.lo)
		testDiv("Div64", bits.Div64, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div64 symmetric", bits.Div64, a.hi, a.lo+a.r, a.x, a.y, a.r)
		// The above code can't test intrinsic implementation, because the passed function is not called directly.
		// The following code uses a closure to test the intrinsic version in case the function is intrinsified.
		testMul("Mul64 intrinsic", func(x, y uint64) (uint64, uint64) { return bits.Mul64(x, y) }, a.x, a.y, a.hi, a.lo)
		testMul("Mul64 intrinsic symmetric", func(x, y uint64) (uint64, uint64) { return bits.Mul64(x, y) }, a.y, a.x, a.hi, a.lo)
		testDiv("Div64 intrinsic", func(hi, lo, y uint64) (uint64, uint64) { return bits.Div64(hi, lo, y) }, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div64 intrinsic symmetric", func(hi, lo, y uint64) (uint64, uint64) { return bits.Div64(hi, lo, y) }, a.hi, a.lo+a.r, a.x, a.y, a.r)
	}
}

func TestDivPanicOverflow(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div should have panicked when y<=hi")
		} else if err != bits.OverflowError {
			t.Errorf("Div expected panic: %q, got: %v ", bits.OverflowError, err)
		}
	}()
	q, r := bits.Div(1, 0, 1)
	t.Errorf("undefined q, r = %v, %v calculated when Div should have panicked", q, r)
}

func TestDiv32PanicOverflow(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div32 should have panicked when y<=hi")
		} else if err != bits.OverflowError {
			t.Errorf("Div32 expected panic: %q, got: %v ", bits.OverflowError, err)
		}
	}()
	q, r := bits.Div32(1, 0, 1)
	t.Errorf("undefined q, r = %v, %v calculated when Div32 should have panicked", q, r)
}

func TestDiv64PanicOverflow(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div64 should have panicked when y<=hi")
		} else if err != bits.OverflowError {
			t.Errorf("Div64 expected panic: %q, got: %v ", bits.OverflowError, err)
		}
	}()
	q, r := bits.Div64(1, 0, 1)
	t.Errorf("undefined q, r = %v, %v calculated when Div64 should have panicked", q, r)
}

func TestDivPanicZero(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div should have panicked when y==0")
		} else if err != bits.DivideError {
			t.Errorf("Div expected panic: %q, got: %q ", bits.DivideError, err)
		}
	}()
	q, r := bits.Div(1, 1, 0)
	t.Errorf("undefined q, r = %v, %v calculated when Div should have panicked", q, r)
}

func TestDiv32PanicZero(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div32 should have panicked when y==0")
		} else if err != bits.DivideError {
			t.Errorf("Div32 expected panic: %q, got: %q ", bits.DivideError, err)
		}
	}()
	q, r := bits.Div32(1, 1, 0)
	t.Errorf("undefined q, r = %v, %v calculated when Div32 should have panicked", q, r)
}

func TestDiv64PanicZero(t *testing.T) {
	// Expect a panic
	defer func() {
		if err := recover(); err == nil {
			t.Error("Div64 should have panicked when y==0")
		} else if err != bits.DivideError {
			t.Errorf("Div64 expected panic: %q, got: %q ", bits.DivideError, err)
		}
	}()
	q, r := bits.Div64(1, 1, 0)
	t.Errorf("undefined q, r = %v, %v calculated when Div64 should have panicked", q, r)
}

func TestRem32(t *testing.T) {
	// Sanity check: for non-overflowing dividends, the result is the
	// same as the rem returned by Div32
	hi, lo, y := uint32(510510), uint32(9699690), uint32(510510+1) // ensure hi < y
	for i := 0; i < 1000; i++ {
		r := bits.Rem32(hi, lo, y)
		_, r2 := bits.Div32(hi, lo, y)
		if r != r2 {
			t.Errorf("Rem32(%v, %v, %v) returned %v, but Div32 returned rem %v", hi, lo, y, r, r2)
		}
		y += 13
	}
}

func TestRem32Overflow(t *testing.T) {
	// To trigger a quotient overflow, we need y <= hi
	hi, lo, y := uint32(510510), uint32(9699690), uint32(7)
	for i := 0; i < 1000; i++ {
		r := bits.Rem32(hi, lo, y)
		_, r2 := bits.Div64(0, uint64(hi)<<32|uint64(lo), uint64(y))
		if r != uint32(r2) {
			t.Errorf("Rem32(%v, %v, %v) returned %v, but Div64 returned rem %v", hi, lo, y, r, r2)
		}
		y += 13
	}
}

func TestRem64(t *testing.T) {
	// Sanity check: for non-overflowing dividends, the result is the
	// same as the rem returned by Div64
	hi, lo, y := uint64(510510), uint64(9699690), uint64(510510+1) // ensure hi < y
	for i := 0; i < 1000; i++ {
		r := bits.Rem64(hi, lo, y)
		_, r2 := bits.Div64(hi, lo, y)
		if r != r2 {
			t.Errorf("Rem64(%v, %v, %v) returned %v, but Div64 returned rem %v", hi, lo, y, r, r2)
		}
		y += 13
	}
}

func TestRem64Overflow(t *testing.T) {
	Rem64Tests := []struct {
		hi, lo, y uint64
		rem       uint64
	}{
		// Testcases computed using Python 3, as:
		//   >>> hi = 42; lo = 1119; y = 42
		//   >>> ((hi<<64)+lo) % y
		{42, 1119, 42, 27},
		{42, 1119, 38, 9},
		{42, 1119, 26, 23},
		{469, 0, 467, 271},
		{469, 0, 113, 58},
		{111111, 111111, 1171, 803},
		{3968194946088682615, 3192705705065114702, 1000037, 56067},
	}

	for _, rt := range Rem64Tests {
		if rt.hi < rt.y {
			t.Fatalf("Rem64(%v, %v, %v) is not a test with quo overflow", rt.hi, rt.lo, rt.y)
		}
		rem := bits.Rem64(rt.hi, rt.lo, rt.y)
		if rem != rt.rem {
			t.Errorf("Rem64(%v, %v, %v) returned %v, wanted %v",
				rt.hi, rt.lo, rt.y, rem, rt.rem)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	var z, c uint
	for i := 0; i < b.N; i++ {
		z, c = bits.Add(uint(Input), uint(i), c)
	}
	Output = int(z + c)
}

func BenchmarkAdd32(b *testing.B) {
	var z, c uint32
	for i := 0; i < b.N; i++ {
		z, c = bits.Add32(uint32(Input), uint32(i), c)
	}
	Output = int(z + c)
}

func BenchmarkAdd64(b *testing.B) {
	var z, c uint64
	for i := 0; i < b.N; i++ {
		z, c = bits.Add64(uint64(Input), uint64(i), c)
	}
	Output = int(z + c)
}

func BenchmarkAdd64multiple(b *testing.B) {
	z0 := uint64(Input)
	z1 := uint64(Input)
	z2 := uint64(Input)
	z3 := uint64(Input)
	for i := 0; i < b.N; i++ {
		var c uint64
		z0, c = bits.Add64(z0, uint64(i), c)
		z1, c = bits.Add64(z1, uint64(i), c)
		z2, c = bits.Add64(z2, uint64(i), c)
		z3, _ = bits.Add64(z3, uint64(i), c)
	}
	Output = int(z0 + z1 + z2 + z3)
}

func BenchmarkSub(b *testing.B) {
	var z, c uint
	for i := 0; i < b.N; i++ {
		z, c = bits.Sub(uint(Input), uint(i), c)
	}
	Output = int(z + c)
}

func BenchmarkSub32(b *testing.B) {
	var z, c uint32
	for i := 0; i < b.N; i++ {
		z, c = bits.Sub32(uint32(Input), uint32(i), c)
	}
	Output = int(z + c)
}

func BenchmarkSub64(b *testing.B) {
	var z, c uint64
	for i := 0; i < b.N; i++ {
		z, c = bits.Sub64(uint64(Input), uint64(i), c)
	}
	Output = int(z + c)
}

func BenchmarkSub64multiple(b *testing.B) {
	z0 := uint64(Input)
	z1 := uint64(Input)
	z2 := uint64(Input)
	z3 := uint64(Input)
	for i := 0; i < b.N; i++ {
		var c uint64
		z0, c = bits.Sub64(z0, uint64(i), c)
		z1, c = bits.Sub64(z1, uint64(i), c)
		z2, c = bits.Sub64(z2, uint64(i), c)
		z3, _ = bits.Sub64(z3, uint64(i), c)
	}
	Output = int(z0 + z1 + z2 + z3)
}

func BenchmarkMul(b *testing.B) {
	var hi, lo uint
	for i := 0; i < b.N; i++ {
		hi, lo = bits.Mul(uint(Input), uint(i))
	}
	Output = int(hi + lo)
}

func BenchmarkMul32(b *testing.B) {
	var hi, lo uint32
	for i := 0; i < b.N; i++ {
		hi, lo = bits.Mul32(uint32(Input), uint32(i))
	}
	Output = int(hi + lo)
}

func BenchmarkMul64(b *testing.B) {
	var hi, lo uint64
	for i := 0; i < b.N; i++ {
		hi, lo = bits.Mul64(uint64(Input), uint64(i))
	}
	Output = int(hi + lo)
}

func BenchmarkDiv(b *testing.B) {
	var q, r uint
	for i := 0; i < b.N; i++ {
		q, r = bits.Div(1, uint(i), uint(Input))
	}
	Output = int(q + r)
}

func BenchmarkDiv32(b *testing.B) {
	var q, r uint32
	for i := 0; i < b.N; i++ {
		q, r = bits.Div32(1, uint32(i), uint32(Input))
	}
	Output = int(q + r)
}

func BenchmarkDiv64(b *testing.B) {
	var q, r uint64
	for i := 0; i < b.N; i++ {
		q, r = bits.Div64(1, uint64(i), uint64(Input))
	}
	Output = int(q + r)
}

// ----------------------------------------------------------------------------
// Testing support

type entry = struct {
	nlz, ntz, pop int
}

// tab contains results for all uint8 values
var tab [256]entry

func init() {
	tab[0] = entry{8, 8, 0}
	for i := 1; i < len(tab); i++ {
		// nlz
		x := i // x != 0
		n := 0
		for x&0x80 == 0 {
			n++
			x <<= 1
		}
		tab[i].nlz = n

		// ntz
		x = i // x != 0
		n = 0
		for x&1 == 0 {
			n++
			x >>= 1
		}
		tab[i].ntz = n

		// pop
		x = i // x != 0
		n = 0
		for x != 0 {
			n += int(x & 1)
			x >>= 1
		}
		tab[i].pop = n
	}
}

// -----------------------------------------------------------------------------
// Tests ported from upstream go1.25.9 example_test.go and example_math_test.go.
//
// Gno's test harness does not dispatch Example* funcs, so each upstream
// ExampleX is rewritten as TestExampleX asserting formatted output matches
// the upstream "// Output: ..." comment.
//
// TestUintSize is adapted: upstream uses unsafe.Sizeof; here we just check
// that UintSize matches Gno's invariant of 64-bit uint.
// -----------------------------------------------------------------------------

func checkExample(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s output mismatch\n got: %q\nwant: %q (got bytes: % x)",
			name, got, want, []byte(got))
	}
}

// TestUintSize: Gno is always 64-bit. Upstream uses unsafe.Sizeof(uint).
func TestUintSize(t *testing.T) {
	if bits.UintSize != 64 {
		t.Fatalf("UintSize = %d; want 64 (Gno is always 64-bit)", bits.UintSize)
	}
}

// --------------------------------------------------------------------
// Examples (upstream example_test.go), rewritten as Test wrappers.
// --------------------------------------------------------------------

func TestExampleLeadingZeros8(t *testing.T) {
	got := fmt.Sprintf("LeadingZeros8(%08b) = %d\n", 1, bits.LeadingZeros8(1))
	checkExample(t, "LeadingZeros8", got, "LeadingZeros8(00000001) = 7\n")
}

func TestExampleLeadingZeros16(t *testing.T) {
	got := fmt.Sprintf("LeadingZeros16(%016b) = %d\n", 1, bits.LeadingZeros16(1))
	checkExample(t, "LeadingZeros16", got, "LeadingZeros16(0000000000000001) = 15\n")
}

func TestExampleLeadingZeros32(t *testing.T) {
	got := fmt.Sprintf("LeadingZeros32(%032b) = %d\n", 1, bits.LeadingZeros32(1))
	checkExample(t, "LeadingZeros32", got,
		"LeadingZeros32(00000000000000000000000000000001) = 31\n")
}

func TestExampleLeadingZeros64(t *testing.T) {
	got := fmt.Sprintf("LeadingZeros64(%064b) = %d\n", 1, bits.LeadingZeros64(1))
	checkExample(t, "LeadingZeros64", got,
		"LeadingZeros64(0000000000000000000000000000000000000000000000000000000000000001) = 63\n")
}

func TestExampleTrailingZeros8(t *testing.T) {
	got := fmt.Sprintf("TrailingZeros8(%08b) = %d\n", 14, bits.TrailingZeros8(14))
	checkExample(t, "TrailingZeros8", got, "TrailingZeros8(00001110) = 1\n")
}

func TestExampleTrailingZeros16(t *testing.T) {
	got := fmt.Sprintf("TrailingZeros16(%016b) = %d\n", 14, bits.TrailingZeros16(14))
	checkExample(t, "TrailingZeros16", got, "TrailingZeros16(0000000000001110) = 1\n")
}

func TestExampleTrailingZeros32(t *testing.T) {
	got := fmt.Sprintf("TrailingZeros32(%032b) = %d\n", 14, bits.TrailingZeros32(14))
	checkExample(t, "TrailingZeros32", got,
		"TrailingZeros32(00000000000000000000000000001110) = 1\n")
}

func TestExampleTrailingZeros64(t *testing.T) {
	got := fmt.Sprintf("TrailingZeros64(%064b) = %d\n", 14, bits.TrailingZeros64(14))
	checkExample(t, "TrailingZeros64", got,
		"TrailingZeros64(0000000000000000000000000000000000000000000000000000000000001110) = 1\n")
}

func TestExampleOnesCount(t *testing.T) {
	got := fmt.Sprintf("OnesCount(%b) = %d\n", 14, bits.OnesCount(14))
	checkExample(t, "OnesCount", got, "OnesCount(1110) = 3\n")
}

func TestExampleOnesCount8(t *testing.T) {
	got := fmt.Sprintf("OnesCount8(%08b) = %d\n", 14, bits.OnesCount8(14))
	checkExample(t, "OnesCount8", got, "OnesCount8(00001110) = 3\n")
}

func TestExampleOnesCount16(t *testing.T) {
	got := fmt.Sprintf("OnesCount16(%016b) = %d\n", 14, bits.OnesCount16(14))
	checkExample(t, "OnesCount16", got, "OnesCount16(0000000000001110) = 3\n")
}

func TestExampleOnesCount32(t *testing.T) {
	got := fmt.Sprintf("OnesCount32(%032b) = %d\n", 14, bits.OnesCount32(14))
	checkExample(t, "OnesCount32", got,
		"OnesCount32(00000000000000000000000000001110) = 3\n")
}

func TestExampleOnesCount64(t *testing.T) {
	got := fmt.Sprintf("OnesCount64(%064b) = %d\n", 14, bits.OnesCount64(14))
	checkExample(t, "OnesCount64", got,
		"OnesCount64(0000000000000000000000000000000000000000000000000000000000001110) = 3\n")
}

func TestExampleRotateLeft8(t *testing.T) {
	got := fmt.Sprintf("%08b\n%08b\n%08b\n",
		15, bits.RotateLeft8(15, 2), bits.RotateLeft8(15, -2))
	checkExample(t, "RotateLeft8", got, "00001111\n00111100\n11000011\n")
}

func TestExampleRotateLeft16(t *testing.T) {
	got := fmt.Sprintf("%016b\n%016b\n%016b\n",
		15, bits.RotateLeft16(15, 2), bits.RotateLeft16(15, -2))
	checkExample(t, "RotateLeft16", got,
		"0000000000001111\n0000000000111100\n1100000000000011\n")
}

func TestExampleRotateLeft32(t *testing.T) {
	got := fmt.Sprintf("%032b\n%032b\n%032b\n",
		15, bits.RotateLeft32(15, 2), bits.RotateLeft32(15, -2))
	checkExample(t, "RotateLeft32", got,
		"00000000000000000000000000001111\n00000000000000000000000000111100\n11000000000000000000000000000011\n")
}

func TestExampleRotateLeft64(t *testing.T) {
	got := fmt.Sprintf("%064b\n%064b\n%064b\n",
		15, bits.RotateLeft64(15, 2), bits.RotateLeft64(15, -2))
	want := "0000000000000000000000000000000000000000000000000000000000001111\n" +
		"0000000000000000000000000000000000000000000000000000000000111100\n" +
		"1100000000000000000000000000000000000000000000000000000000000011\n"
	checkExample(t, "RotateLeft64", got, want)
}

func TestExampleReverse8(t *testing.T) {
	got := fmt.Sprintf("%08b\n%08b\n", 19, bits.Reverse8(19))
	checkExample(t, "Reverse8", got, "00010011\n11001000\n")
}

func TestExampleReverse16(t *testing.T) {
	got := fmt.Sprintf("%016b\n%016b\n", 19, bits.Reverse16(19))
	checkExample(t, "Reverse16", got, "0000000000010011\n1100100000000000\n")
}

func TestExampleReverse32(t *testing.T) {
	got := fmt.Sprintf("%032b\n%032b\n", 19, bits.Reverse32(19))
	checkExample(t, "Reverse32", got,
		"00000000000000000000000000010011\n11001000000000000000000000000000\n")
}

func TestExampleReverse64(t *testing.T) {
	got := fmt.Sprintf("%064b\n%064b\n", 19, bits.Reverse64(19))
	want := "0000000000000000000000000000000000000000000000000000000000010011\n" +
		"1100100000000000000000000000000000000000000000000000000000000000\n"
	checkExample(t, "Reverse64", got, want)
}

func TestExampleReverseBytes16(t *testing.T) {
	got := fmt.Sprintf("%016b\n%016b\n", 15, bits.ReverseBytes16(15))
	checkExample(t, "ReverseBytes16", got, "0000000000001111\n0000111100000000\n")
}

func TestExampleReverseBytes32(t *testing.T) {
	got := fmt.Sprintf("%032b\n%032b\n", 15, bits.ReverseBytes32(15))
	checkExample(t, "ReverseBytes32", got,
		"00000000000000000000000000001111\n00001111000000000000000000000000\n")
}

func TestExampleReverseBytes64(t *testing.T) {
	got := fmt.Sprintf("%064b\n%064b\n", 15, bits.ReverseBytes64(15))
	want := "0000000000000000000000000000000000000000000000000000000000001111\n" +
		"0000111100000000000000000000000000000000000000000000000000000000\n"
	checkExample(t, "ReverseBytes64", got, want)
}

func TestExampleLen8(t *testing.T) {
	got := fmt.Sprintf("Len8(%08b) = %d\n", 8, bits.Len8(8))
	checkExample(t, "Len8", got, "Len8(00001000) = 4\n")
}

func TestExampleLen16(t *testing.T) {
	got := fmt.Sprintf("Len16(%016b) = %d\n", 8, bits.Len16(8))
	checkExample(t, "Len16", got, "Len16(0000000000001000) = 4\n")
}

func TestExampleLen32(t *testing.T) {
	got := fmt.Sprintf("Len32(%032b) = %d\n", 8, bits.Len32(8))
	checkExample(t, "Len32", got,
		"Len32(00000000000000000000000000001000) = 4\n")
}

func TestExampleLen64(t *testing.T) {
	got := fmt.Sprintf("Len64(%064b) = %d\n", 8, bits.Len64(8))
	checkExample(t, "Len64", got,
		"Len64(0000000000000000000000000000000000000000000000000000000000001000) = 4\n")
}

// --------------------------------------------------------------------
// Examples (upstream example_math_test.go), rewritten as Test wrappers.
// --------------------------------------------------------------------

func TestExampleAdd32(t *testing.T) {
	// 1st number 33<<32 + 12; 2nd 21<<32 + 23; no carry.
	n1 := []uint32{33, 12}
	n2 := []uint32{21, 23}
	d1, carry := bits.Add32(n1[1], n2[1], 0)
	d0, _ := bits.Add32(n1[0], n2[0], carry)
	nsum := []uint32{d0, d1}
	got := fmt.Sprintf("%v + %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	// 1st 1<<32 + 2147483648; 2nd 1<<32 + 2147483648; carry expected.
	n1 = []uint32{1, 0x80000000}
	n2 = []uint32{1, 0x80000000}
	d1, carry = bits.Add32(n1[1], n2[1], 0)
	d0, _ = bits.Add32(n1[0], n2[0], carry)
	nsum = []uint32{d0, d1}
	got += fmt.Sprintf("%v + %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	want := "[33 12] + [21 23] = [54 35] (carry bit was 0)\n" +
		"[1 2147483648] + [1 2147483648] = [3 0] (carry bit was 1)\n"
	checkExample(t, "Add32", got, want)
}

func TestExampleAdd64(t *testing.T) {
	n1 := []uint64{33, 12}
	n2 := []uint64{21, 23}
	d1, carry := bits.Add64(n1[1], n2[1], 0)
	d0, _ := bits.Add64(n1[0], n2[0], carry)
	nsum := []uint64{d0, d1}
	got := fmt.Sprintf("%v + %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	n1 = []uint64{1, 0x8000000000000000}
	n2 = []uint64{1, 0x8000000000000000}
	d1, carry = bits.Add64(n1[1], n2[1], 0)
	d0, _ = bits.Add64(n1[0], n2[0], carry)
	nsum = []uint64{d0, d1}
	got += fmt.Sprintf("%v + %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	want := "[33 12] + [21 23] = [54 35] (carry bit was 0)\n" +
		"[1 9223372036854775808] + [1 9223372036854775808] = [3 0] (carry bit was 1)\n"
	checkExample(t, "Add64", got, want)
}

func TestExampleSub32(t *testing.T) {
	n1 := []uint32{33, 23}
	n2 := []uint32{21, 12}
	d1, carry := bits.Sub32(n1[1], n2[1], 0)
	d0, _ := bits.Sub32(n1[0], n2[0], carry)
	nsum := []uint32{d0, d1}
	got := fmt.Sprintf("%v - %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	n1 = []uint32{3, 0x7fffffff}
	n2 = []uint32{1, 0x80000000}
	d1, carry = bits.Sub32(n1[1], n2[1], 0)
	d0, _ = bits.Sub32(n1[0], n2[0], carry)
	nsum = []uint32{d0, d1}
	got += fmt.Sprintf("%v - %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	want := "[33 23] - [21 12] = [12 11] (carry bit was 0)\n" +
		"[3 2147483647] - [1 2147483648] = [1 4294967295] (carry bit was 1)\n"
	checkExample(t, "Sub32", got, want)
}

func TestExampleSub64(t *testing.T) {
	n1 := []uint64{33, 23}
	n2 := []uint64{21, 12}
	d1, carry := bits.Sub64(n1[1], n2[1], 0)
	d0, _ := bits.Sub64(n1[0], n2[0], carry)
	nsum := []uint64{d0, d1}
	got := fmt.Sprintf("%v - %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	n1 = []uint64{3, 0x7fffffffffffffff}
	n2 = []uint64{1, 0x8000000000000000}
	d1, carry = bits.Sub64(n1[1], n2[1], 0)
	d0, _ = bits.Sub64(n1[0], n2[0], carry)
	nsum = []uint64{d0, d1}
	got += fmt.Sprintf("%v - %v = %v (carry bit was %v)\n", n1, n2, nsum, carry)

	want := "[33 23] - [21 12] = [12 11] (carry bit was 0)\n" +
		"[3 9223372036854775807] - [1 9223372036854775808] = [1 18446744073709551615] (carry bit was 1)\n"
	checkExample(t, "Sub64", got, want)
}

func TestExampleMul32(t *testing.T) {
	n1 := []uint32{0, 12}
	n2 := []uint32{0, 12}
	hi, lo := bits.Mul32(n1[1], n2[1])
	nsum := []uint32{hi, lo}
	got := fmt.Sprintf("%v * %v = %v\n", n1[1], n2[1], nsum)

	n1 = []uint32{0, 0x80000000}
	n2 = []uint32{0, 2}
	hi, lo = bits.Mul32(n1[1], n2[1])
	nsum = []uint32{hi, lo}
	got += fmt.Sprintf("%v * %v = %v\n", n1[1], n2[1], nsum)

	want := "12 * 12 = [0 144]\n2147483648 * 2 = [1 0]\n"
	checkExample(t, "Mul32", got, want)
}

func TestExampleMul64(t *testing.T) {
	n1 := []uint64{0, 12}
	n2 := []uint64{0, 12}
	hi, lo := bits.Mul64(n1[1], n2[1])
	nsum := []uint64{hi, lo}
	got := fmt.Sprintf("%v * %v = %v\n", n1[1], n2[1], nsum)

	n1 = []uint64{0, 0x8000000000000000}
	n2 = []uint64{0, 2}
	hi, lo = bits.Mul64(n1[1], n2[1])
	nsum = []uint64{hi, lo}
	got += fmt.Sprintf("%v * %v = %v\n", n1[1], n2[1], nsum)

	want := "12 * 12 = [0 144]\n9223372036854775808 * 2 = [1 0]\n"
	checkExample(t, "Mul64", got, want)
}

func TestExampleDiv32(t *testing.T) {
	n1 := []uint32{0, 6}
	n2 := []uint32{0, 3}
	quo, rem := bits.Div32(n1[0], n1[1], n2[1])
	nsum := []uint32{quo, rem}
	got := fmt.Sprintf("[%v %v] / %v = %v\n", n1[0], n1[1], n2[1], nsum)

	n1 = []uint32{2, 0x80000000}
	n2 = []uint32{0, 0x80000000}
	quo, rem = bits.Div32(n1[0], n1[1], n2[1])
	nsum = []uint32{quo, rem}
	got += fmt.Sprintf("[%v %v] / %v = %v\n", n1[0], n1[1], n2[1], nsum)

	want := "[0 6] / 3 = [2 0]\n[2 2147483648] / 2147483648 = [5 0]\n"
	checkExample(t, "Div32", got, want)
}

func TestExampleDiv64(t *testing.T) {
	n1 := []uint64{0, 6}
	n2 := []uint64{0, 3}
	quo, rem := bits.Div64(n1[0], n1[1], n2[1])
	nsum := []uint64{quo, rem}
	got := fmt.Sprintf("[%v %v] / %v = %v\n", n1[0], n1[1], n2[1], nsum)

	n1 = []uint64{2, 0x8000000000000000}
	n2 = []uint64{0, 0x8000000000000000}
	quo, rem = bits.Div64(n1[0], n1[1], n2[1])
	nsum = []uint64{quo, rem}
	got += fmt.Sprintf("[%v %v] / %v = %v\n", n1[0], n1[1], n2[1], nsum)

	want := "[0 6] / 3 = [2 0]\n[2 9223372036854775808] / 9223372036854775808 = [5 0]\n"
	checkExample(t, "Div64", got, want)
}
