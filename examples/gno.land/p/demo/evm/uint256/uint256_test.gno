package uint256

import "testing"

func TestCtor(t *testing.T) {
	u := New()
	if u.String() != "0" {
		t.Error("expected to be 0")
	}
}

func Test1Ctor64(t *testing.T) {
	u := NewFrom64s(42)
	if u.String() != "42" {
		t.Error("expected to be 42")
	}
}

func Test2Ctor64(t *testing.T) {
	u := NewFrom64s(42, 1)
	if u.String() != "18446744073709551658" {
		t.Error("expected to be 18446744073709551658")
	}
}

func TestAdd(t *testing.T) {
	u := New()
	u.Add64(42)
	if u.String() != "42" {
		t.Error("expected to be 42")
	}
}

func TestMul(t *testing.T) {
	u := New()
	u.Add64(21)
	u.Mul64(2)
	if u.String() != "42" {
		t.Error("expected to be 42")
	}
}

func TestFromString(t *testing.T) {
	u, err := NewFromString("ffffffffffffffff", 16)
	if err != nil {
		t.Error(err)
	}
	if u.String() != "18446744073709551615" {
		t.Error("expected to be 18446744073709551615")
	}
}

func TestMulLevel(t *testing.T) {
	u, _ := NewFromString("ffffffffffffffff", 16)
	u.Mul64(2)
	if u.data[1] != 1 {
		t.Error("expected second data to be 1")
	}
	if u.String() != "36893488147419103230" {
		t.Error("expected to be 36893488147419103230")
	}
}

func TestStringBase(t *testing.T) {
	u, _ := NewFromString("100010101010101000100101110101", 2)
	if u.String() != "581601653" {
		t.Error("expected to be 581601653")
	}
}

func TestLastLevel(t *testing.T) {
	u, err := NewFromString("ffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	if err != nil {
		t.Error(err)
	}
	if u.String() != "6277101735386680763835789423207666416102355444464034512895" {
		t.Error("expected to be 6277101735386680763835789423207666416102355444464034512895")
	}
	u.Add64(1)
	if u.String() != "6277101735386680763835789423207666416102355444464034512896" {
		t.Error("expected to be 6277101735386680763835789423207666416102355444464034512896")
	}
	if u.data[0] != 0 || u.data[1] != 0 || u.data[2] != 0 || u.data[3] != 1 {
		t.Error("expected [0, 0, 0, 1]")
	}
	u.Add64(1)
	if u.String() != "6277101735386680763835789423207666416102355444464034512897" {
		t.Error("expected to be 6277101735386680763835789423207666416102355444464034512897")
	}
	if u.data[0] != 1 || u.data[1] != 0 || u.data[2] != 0 || u.data[3] != 1 {
		t.Error("expected [1, 0, 0, 1]")
	}
}