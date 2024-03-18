package uint256

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"gno.land/p/demo/ufmt"
	i256 "gno.land/p/demo/int256"
)

// TODO: Set up simple tests for now and modify them later to check for more complex cases.

func parseUintFromString(s string) (*Uint, error) {
	if s == "0" {
		return NewUint(0), nil
	}
	value, err := parseUint(s, 16, 64)
	if err != nil {
		return nil, err
	}
	return NewUint(value), nil
}

type binOpTest struct {
	x, y, want string
}

// TODO: use `SetBytes` to test hex strings
func TestAdd(t *testing.T) {
	tests := []binOpTest{
		{"0", "1", "1"},
		{"1", "0", "1"},
		{"1", "1", "2"},
		{"1", "3", "4"},
		{"10", "10", "20"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := parseUintFromString(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Add(x, y)

		for i := range got.arr {
			if got.arr[i] != want.arr[i] {
				t.Errorf("Add(%s, %s) = %v, want %v", tc.x, tc.y, got, want)
				break
			}
		}
	}
}

// TODO: need to test negative results
func TestSub(t *testing.T) {
	tests := []binOpTest{
		{"1", "0", "1"},
		{"1", "1", "0"},
		{"10", "10", "0"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := parseUintFromString(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Sub(x, y)

		for i := range got.arr {
			if got.arr[i] != want.arr[i] {
				t.Errorf("Sub(%s, %s) = %v, want %v", tc.x, tc.y, got, want)
				break
			}
		}
	}
}

func TestMul(t *testing.T) {
	tests := []binOpTest{
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"10", "10", "100"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		want, err := parseUintFromString(tc.want)
		if err != nil {
			t.Error(err)
			continue
		}

		got := &Uint{}
		got.Mul(x, y)

		for i := range got.arr {
			if got.arr[i] != want.arr[i] {
				t.Errorf("Mul(%s, %s) = %v, want %v", tc.x, tc.y, got, want)
				break
			}
		}
	}
}

func TestDivMod(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		wantDiv string
		wantMod string
	}{
		{"1", "1", "1", "0"},
		{"10", "10", "1", "0"},
		{"100", "10", "10", "0"},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		wantDiv, err := parseUintFromString(tc.wantDiv)
		if err != nil {
			t.Error(err)
			continue
		}

		wantMod, err := parseUintFromString(tc.wantMod)
		if err != nil {
			t.Error(err)
			continue
		}

		gotDiv := new(Uint)
		gotMod := new(Uint)
		gotDiv.DivMod(x, y, gotMod)

		for i := range gotDiv.arr {
			if gotDiv.arr[i] != wantDiv.arr[i] {
				t.Errorf("DivMod(%s, %s) got Div %v, want Div %v", tc.x, tc.y, gotDiv, wantDiv)
				break
			}
		}
		for i := range gotMod.arr {
			if gotMod.arr[i] != wantMod.arr[i] {
				t.Errorf("DivMod(%s, %s) got Mod %v, want Mod %v", tc.x, tc.y, gotMod, wantMod)
				break
			}
		}
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		x    string
		want bool
	}{
		{"0", true},
		{"1", false},
		{"10", false},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.IsZero()
		if got != tc.want {
			t.Errorf("IsZero(%s) = %v, want %v", tc.x, got, tc.want)
		}
	}
}

func TestIsUint64(t *testing.T) {
	tests := []struct {
		x    string
		want bool
	}{
		{"0", true},
		{"1", true},
		{"10", true},
		{"ffffffffffffffff", true},
		{"10000000000000000", true},
	}

	for _, tc := range tests {
		x := new(Uint).SetBytes(hex2Bytes(tc.x))
		got := x.IsUint64()

		if got != tc.want {
			t.Errorf("IsUint64(%s) = %v, want %v", tc.x, got, tc.want)
		}
	}
}

func TestLtUint64(t *testing.T) {
	tests := []struct {
		x    string
		y    uint64
		want bool
	}{
		{"0", 1, true},
		{"1", 0, false},
		{"10", 10, false},
		{"ffffffffffffffff", 0, false},
		{"10000000000000000", 0, false},
	}

	for _, tc := range tests {
		x := new(Uint).SetBytes(hex2Bytes(tc.x))
		got := x.LtUint64(tc.y)

		if got != tc.want {
			t.Errorf("LtUint64(%s, %d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestCmp(t *testing.T) {
	tests := []struct {
		x, y string
		want int
	}{
		{"0", "0", 0},
		{"0", "1", -1},
		{"1", "0", 1},
		{"1", "1", 0},
		{"10", "10", 0},
		{"10", "11", -1},
		{"11", "10", 1},
	}

	for _, tc := range tests {
		x, err := parseUintFromString(tc.x)
		if err != nil {
			t.Error(err)
			continue
		}

		y, err := parseUintFromString(tc.y)
		if err != nil {
			t.Error(err)
			continue
		}

		got := x.Cmp(y)
		if got != tc.want {
			t.Errorf("Cmp(%s, %s) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

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

func TestSGT(t *testing.T) {
	x := new(Uint).SetBytes(hex2Bytes("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe"))
	y := new(Uint).SetBytes(hex2Bytes("00"))
	actual := x.Sgt(y)
	if actual {
		t.Fatalf("Expected %v false", actual)
	}

	x = new(Uint).SetBytes(hex2Bytes("00"))
	y = new(Uint).SetBytes(hex2Bytes("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe"))
	actual = x.Sgt(y)
	if !actual {
		t.Fatalf("Expected %v true", actual)
	}
}

type logicOpTest struct {
	x, y, want string
}
