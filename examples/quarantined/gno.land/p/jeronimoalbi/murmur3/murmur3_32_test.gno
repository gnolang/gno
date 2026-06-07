package murmur3_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/murmur3"
)

func TestHash32Sum32(t *testing.T) {
	// Test cases -> https://en.wikipedia.org/wiki/MurmurHash#Algorithm
	tests := []struct {
		name string
		data string
		seed uint32
		want uint32
	}{
		{"empty with seed 0", "", 0, 0},
		{"empty with seed 1", "", 1, 0x514e28b7},
		{"empty with seed max", "", 0xffffffff, 0x81f16f39},
		{"test", "test", 0, 0xba6bd213},
		{"test with seed", "test", 0x9747b28c, 0x704b81dc},
		{"hello", "Hello, world!", 0, 0xc0363e43},
		{"hello with seed", "Hello, world!", 42, 0x2c8c8533},
		{"hello with seed 0x9747b28c", "Hello, world!", 0x9747b28c, 0x24884cba},
		{"pangram", "The quick brown fox jumps over the lazy dog", 0, 0x2e4ff723},
		{"pangram with seed", "The quick brown fox jumps over the lazy dog", 0x9747b28c, 0x2fa826cd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := murmur3.Sum32WithSeed([]byte(tt.data), tt.seed)
			if got != tt.want {
				t.Errorf("Sum32WithSeed(%q, %d) = 0x%08x, want 0x%08x", tt.data, tt.seed, got, tt.want)
			}
		})
	}
}

func TestHash32InterfaceSizes(t *testing.T) {
	h := murmur3.New32()
	if h.Size() != 4 {
		t.Errorf("Size() = %d, want 4", h.Size())
	}
	if h.BlockSize() != 4 {
		t.Errorf("BlockSize() = %d, want 4", h.BlockSize())
	}
}

func TestHash32IncrementalWrite(t *testing.T) {
	want := uint32(3224780355)
	data := []byte("Hello, world!")

	// Hash one byte at a time
	h := murmur3.New32()
	for _, b := range data {
		h.Write([]byte{b})
	}

	if got := h.Sum32(); got != want {
		t.Errorf("incremental Write got 0x%08x, want 0x%08x", got, want)
	}

	// Hash in two parts
	h.Reset()
	h.Write(data[:5])
	h.Write(data[5:])
	if got := h.Sum32(); got != want {
		t.Errorf("two-part Write got 0x%08x, want 0x%08x", got, want)
	}
}

func TestHash32Sum(t *testing.T) {
	h := murmur3.New32()
	h.Write([]byte("Hello, world!"))
	b := h.Sum(nil)
	if len(b) != 4 {
		t.Fatalf("Sum(nil) returned %d bytes, want 4", len(b))
	}

	got := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	want := uint32(3224780355)
	if got != want {
		t.Errorf("Sum got 0x%08x, want 0x%08x", got, want)
	}
}

func TestHash32SumDoesNotChangeState(t *testing.T) {
	h := murmur3.New32()
	h.Write([]byte("Hello"))
	s1 := h.Sum32()
	s2 := h.Sum32()
	if s1 != s2 {
		t.Errorf("Sum32 changed state: first 0x%08x, second 0x%08x", s1, s2)
	}
}

func TestHash32Reset(t *testing.T) {
	h := murmur3.New32()
	h.Write([]byte("some data"))
	h.Reset()
	h.Write([]byte("Hello, world!"))
	got := h.Sum32()
	want := uint32(3224780355)
	if got != want {
		t.Errorf("after Reset got 0x%08x, want 0x%08x", got, want)
	}
}
