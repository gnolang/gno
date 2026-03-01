package ternary

import (
	"std"
	"testing"
)

func TestTernary(t *testing.T) {
	assert(t, String(true, "a", "b") == "a")
	assert(t, String(false, "a", "b") == "b")
	assert(t, Int(true, 0, 1) == 0)
	assert(t, Int(false, 0, 1) == 1)
	assert(t, Int8(true, 0, 1) == 0)
	assert(t, Int8(false, 0, 1) == 1)
	assert(t, Int16(true, 0, 1) == 0)
	assert(t, Int16(false, 0, 1) == 1)
	assert(t, Int32(true, 0, 1) == 0)
	assert(t, Int32(false, 0, 1) == 1)
	assert(t, Int64(true, 0, 1) == 0)
	assert(t, Int64(false, 0, 1) == 1)
	assert(t, Uint(true, 0, 1) == 0)
	assert(t, Uint(false, 0, 1) == 1)
	assert(t, Uint8(true, 0, 1) == 0)
	assert(t, Uint8(false, 0, 1) == 1)
	assert(t, Uint16(true, 0, 1) == 0)
	assert(t, Uint16(false, 0, 1) == 1)
	assert(t, Uint32(true, 0, 1) == 0)
	assert(t, Uint32(false, 0, 1) == 1)
	assert(t, Uint64(true, 0, 1) == 0)
	assert(t, Uint64(false, 0, 1) == 1)
	assert(t, Float32(true, 3.14, 1.618) == 3.14)
	assert(t, Float32(false, 3.14, 1.618) == 1.618)
	assert(t, Float64(true, 3.14, 1.618) == 3.14)
	assert(t, Float64(false, 3.14, 1.618) == 1.618)
	assert(t, Rune(true, '是', '否') == '是')
	assert(t, Rune(false, '是', '否') == '否')
	n := 17
	assert(t, !Bool(true, n%2 == 0, n < 10))
	assert(t, Address(true, std.Address("g0"), std.Address("g1")).String() == "g0")
	assert(t, Address(false, std.Address("g0"), std.Address("g1")).String() == "g1")
}

func assert(t *testing.T, val bool) {
	t.Helper()
	if !val {
		t.Errorf("expected true, got false")
	}
}
