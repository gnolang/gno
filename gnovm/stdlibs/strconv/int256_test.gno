package strconv_test

import (
	"errors"
	"math"
	"strconv"
	"testing"
)

// An int256 constant must fit an int64, so wide int256 test values are
// built at run time.
var (
	maxBits    uint256 = math.MaxInt256
	i256Max            = int256(maxBits)
	i256Min            = int256(maxBits + 1)
	minInt64   int256  = math.MinInt64
	minInt64m1         = minInt64 - 1
)

const (
	maxUint256Str = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
	maxInt256Str  = "57896044618658097711785492504343953926634992332820282019728792003956564819967"
	minInt256Str  = "-57896044618658097711785492504343953926634992332820282019728792003956564819968"
)

func TestFormatUint256(t *testing.T) {
	tests := []struct {
		in   uint256
		base int
		out  string
	}{
		{0, 10, "0"},
		{1, 10, "1"},
		{math.MaxUint64, 10, "18446744073709551615"},
		{math.MaxUint64 + 1, 10, "18446744073709551616"},
		{1e19, 10, "10000000000000000000"},
		{1e38, 10, "100000000000000000000000000000000000000"},
		{1e38 + 7, 10, "100000000000000000000000000000000000007"},
		{math.MaxUint256, 10, maxUint256Str},
		{math.MaxUint256, 16, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
		{1 << 255, 2, "1" + zeros(255)},
		{1 << 128, 16, "1" + zeros(32)},
		{math.MaxUint256, 36, "6dp5qcb22im238nr3wvp0ic7q99w035jmy2iw7i6n43d37jtof"},
	}
	for _, tt := range tests {
		if got := strconv.FormatUint256(tt.in, tt.base); got != tt.out {
			t.Errorf("FormatUint256(%d, %d) = %q, want %q", tt.in, tt.base, got, tt.out)
		}
		if got := string(strconv.AppendUint256([]byte("x"), tt.in, tt.base)); got != "x"+tt.out {
			t.Errorf("AppendUint256(%d, %d) = %q, want %q", tt.in, tt.base, got, "x"+tt.out)
		}
	}
}

func TestFormatInt256(t *testing.T) {
	tests := []struct {
		in   int256
		base int
		out  string
	}{
		{0, 10, "0"},
		{-1, 10, "-1"},
		{math.MinInt64, 10, "-9223372036854775808"},
		{minInt64m1, 10, "-9223372036854775809"},
		{i256Max, 10, maxInt256Str},
		{i256Min, 10, minInt256Str},
		{i256Min, 16, "-8" + zeros(63)},
		{-255, 16, "-ff"},
	}
	for _, tt := range tests {
		if got := strconv.FormatInt256(tt.in, tt.base); got != tt.out {
			t.Errorf("FormatInt256(%d, %d) = %q, want %q", tt.in, tt.base, got, tt.out)
		}
		if got := string(strconv.AppendInt256(nil, tt.in, tt.base)); got != tt.out {
			t.Errorf("AppendInt256(%d, %d) = %q, want %q", tt.in, tt.base, got, tt.out)
		}
	}
}

func TestFormat256BadBase(t *testing.T) {
	for _, base := range []int{-1, 0, 1, 37} {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("FormatUint256 with base %d: expected panic", base)
				}
			}()
			strconv.FormatUint256(1, base)
		}()
	}
}

func TestParseUint256(t *testing.T) {
	tests := []struct {
		in   string
		base int
		out  uint256
		err  error
	}{
		{"", 10, 0, strconv.ErrSyntax},
		{"0", 10, 0, nil},
		{"1", 10, 1, nil},
		{"18446744073709551616", 10, math.MaxUint64 + 1, nil},
		{maxUint256Str, 10, math.MaxUint256, nil},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639936", 10, math.MaxUint256, strconv.ErrRange},
		{maxUint256Str + "0", 10, math.MaxUint256, strconv.ErrRange},
		{"-1", 10, 0, strconv.ErrSyntax},
		{"+1", 10, 0, strconv.ErrSyntax},
		{"12a", 10, 0, strconv.ErrSyntax},
		{"0x" + "f", 0, 15, nil},
		{"0x1_0", 0, 16, nil},
		{"1_0", 10, 0, strconv.ErrSyntax},
		{"_10", 0, 0, strconv.ErrSyntax},
		{"0b101", 0, 5, nil},
		{"0o17", 0, 15, nil},
		{"017", 0, 15, nil},
		{"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16, math.MaxUint256, nil},
		{"1" + zeros(64), 16, math.MaxUint256, strconv.ErrRange},
		{"1" + zeros(255), 2, 1 << 255, nil},
	}
	for _, tt := range tests {
		out, err := strconv.ParseUint256(tt.in, tt.base)
		if out != tt.out || !errors.Is(err, tt.err) {
			t.Errorf("ParseUint256(%q, %d) = %d, %v; want %d, %v",
				tt.in, tt.base, out, err, tt.out, tt.err)
		}
	}
	if _, err := strconv.ParseUint256("1", 1); err == nil {
		t.Errorf("ParseUint256 with base 1: expected an error")
	}
}

func TestParseInt256(t *testing.T) {
	tests := []struct {
		in  string
		out int256
		err error
	}{
		{"", 0, strconv.ErrSyntax},
		{"0", 0, nil},
		{"-0", 0, nil},
		{"+7", 7, nil},
		{"-7", -7, nil},
		{maxInt256Str, i256Max, nil},
		{minInt256Str, i256Min, nil},
		{"57896044618658097711785492504343953926634992332820282019728792003956564819968", i256Max, strconv.ErrRange},
		{"-57896044618658097711785492504343953926634992332820282019728792003956564819969", i256Min, strconv.ErrRange},
		{maxUint256Str + "1", i256Max, strconv.ErrRange},
		{"-" + maxUint256Str + "1", i256Min, strconv.ErrRange},
		{"--1", 0, strconv.ErrSyntax},
		{"1x", 0, strconv.ErrSyntax},
	}
	for _, tt := range tests {
		out, err := strconv.ParseInt256(tt.in, 10)
		if out != tt.out || !errors.Is(err, tt.err) {
			t.Errorf("ParseInt256(%q) = %d, %v; want %d, %v",
				tt.in, out, err, tt.out, tt.err)
		}
		if err != nil {
			if ne, ok := err.(*strconv.NumError); !ok || ne.Func != "ParseInt256" || ne.Num != tt.in {
				t.Errorf("ParseInt256(%q): unexpected error %#v", tt.in, err)
			}
		}
	}
}

func TestRoundTrip256(t *testing.T) {
	u := uint256(1)
	for i := 0; i < 256; i++ {
		for _, v := range []uint256{u, u - 1, u + 1, ^u} {
			for _, base := range []int{2, 7, 10, 16, 36} {
				s := strconv.FormatUint256(v, base)
				got, err := strconv.ParseUint256(s, base)
				if err != nil || got != v {
					t.Fatalf("round trip %d base %d: %q -> %d, %v", v, base, s, got, err)
				}
				iv := int256(v)
				s = strconv.FormatInt256(iv, base)
				igot, err := strconv.ParseInt256(s, base)
				if err != nil || igot != iv {
					t.Fatalf("round trip %d base %d: %q -> %d, %v", iv, base, s, igot, err)
				}
			}
		}
		u <<= 1
	}
}

func zeros(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '0'
	}
	return string(b)
}
