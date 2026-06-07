package murmur3_test

import (
	"testing"

	"gno.land/p/jeronimoalbi/murmur3"
)

func TestHash64Sum64(t *testing.T) {
	// Test cases -> https://en.wikipedia.org/wiki/MurmurHash#Algorithm
	tests := []struct {
		name string
		data string
		want uint64
	}{
		{"empty", "", 0x00000000514e28b7},
		{"test", "test", 0xba6bd21399c02ae2},
		{"hello", "Hello, world!", 0xc0363e43aa5dc85b},
		{"pangram", "The quick brown fox jumps over the lazy dog", 0x2e4ff72378e69e27},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := murmur3.Sum64([]byte(tt.data))
			if got != tt.want {
				t.Errorf("Sum64(%q) = 0x%016x, want 0x%016x", tt.data, got, tt.want)
			}
		})
	}
}

func TestHash64InterfaceSizes(t *testing.T) {
	h := murmur3.New64()
	if h.Size() != 8 {
		t.Errorf("Size() = %d, want 8", h.Size())
	}
	if h.BlockSize() != 4 {
		t.Errorf("BlockSize() = %d, want 4", h.BlockSize())
	}
}

func TestHash64IncrementalWrite(t *testing.T) {
	want := uint64(13850326164366542939)
	data := []byte("Hello, world!")

	// Hash one byte at a time
	h := murmur3.New64()
	for _, b := range data {
		h.Write([]byte{b})
	}

	if got := h.Sum64(); got != want {
		t.Errorf("incremental Write got 0x%016x, want 0x%016x", got, want)
	}

	// Hash in two parts
	h.Reset()
	h.Write(data[:5])
	h.Write(data[5:])
	if got := h.Sum64(); got != want {
		t.Errorf("two-part Write got 0x%016x, want 0x%016x", got, want)
	}
}

func TestHash64Sum(t *testing.T) {
	h := murmur3.New64()
	h.Write([]byte("Hello, world!"))
	b := h.Sum(nil)
	if len(b) != 8 {
		t.Fatalf("Sum(nil) returned %d bytes, want 8", len(b))
	}

	got := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	want := uint64(13850326164366542939)
	if got != want {
		t.Errorf("Sum got 0x%016x, want 0x%016x", got, want)
	}
}

func TestHash64SumDoesNotChangeState(t *testing.T) {
	h := murmur3.New64()
	h.Write([]byte("Hello"))
	s1 := h.Sum64()
	s2 := h.Sum64()
	if s1 != s2 {
		t.Errorf("Sum64 changed state: first 0x%016x, second 0x%016x", s1, s2)
	}
}

func TestHash64Reset(t *testing.T) {
	h := murmur3.New64()
	h.Write([]byte("some data"))
	h.Reset()
	h.Write([]byte("Hello, world!"))
	got := h.Sum64()
	want := uint64(13850326164366542939)
	if got != want {
		t.Errorf("after Reset got 0x%016x, want 0x%016x", got, want)
	}
}
