package uint256

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"gno.land/p/demo/ufmt"
	i256 "gno.land/p/demo/int256"
)

func hex2Bytes(s string) []byte {
	h, _ := hex.DecodeString(s)
	return h
}

func checkOverflow(f *Uint, overflow bool) error {
	hb := hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	max := NewUint(0).SetBytes(hb)
	shouldOverflow := f.Cmp(max) > 0
	if overflow != shouldOverflow {
		return ufmt.Errorf("overflow should be %v, but got %v", overflow, shouldOverflow)
	}

	return nil
}

func TestByte(t *testing.T) {
	z := new(Uint).SetBytes(hex2Bytes("ABCDEF09080706050403020100000000000000000000000000000000000000ef"))
	actual := z.Byte(NewUint(0))
	expected := new(Uint).SetBytes(hex2Bytes("00000000000000000000000000000000000000000000000000000000000000ab"))
	if !actual.Eq(expected) {
		t.Fatalf("Expected %x, got %x", expected, actual)
	}

	z = new(Uint).SetBytes(hex2Bytes("ABCDEF09080706050403020100000000000000000000000000000000000000ef"))
	actual = z.Byte(NewUint(31))
	expected = new(Uint).SetBytes(hex2Bytes("00000000000000000000000000000000000000000000000000000000000000ef"))
	if !actual.Eq(expected) {
		t.Fatalf("Expected %x, got %x", expected, actual)
	}

	z = new(Uint).SetBytes(hex2Bytes("ABCDEF09080706050403020100000000000000000000000000000000000000ef"))
	actual = z.Byte(NewUint(32))
	expected =new(Uint).SetBytes(hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"))
	if !actual.Eq(expected) {
		t.Fatalf("Expected %x, got %x", expected, actual)
	}

	z = new(Uint).SetBytes(hex2Bytes("ABCDEF0908070605040302011111111111111111111111111111111111111111"))
	actual = z.Byte(new(Uint).SetBytes(hex2Bytes("f000000000000000000000000000000000000000000000000000000000000001")))
	expected = new(Uint).SetBytes(hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"))
	if !actual.Eq(expected) {
		t.Fatalf("Expected %x, got %x", expected, actual)
	}
}
