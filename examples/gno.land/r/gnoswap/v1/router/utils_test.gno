package router

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestGetDataForSinglePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantToken0  string
		wantToken1  string
		wantFee     int
		shouldPanic bool
	}{
		{
			name:        "valid path",
			input:       "tokenA:tokenB:500",
			wantToken0:  "tokenA",
			wantToken1:  "tokenB",
			wantFee:     int(500),
			shouldPanic: false,
		},
		{
			name:        "invalid path format",
			input:       "tokenA:tokenB",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.shouldPanic {
					t.Errorf("getDataForSinglePath() panic = %v, shouldPanic = %v", r != nil, tt.shouldPanic)
				}
			}()

			token0, token1, fee := getDataForSinglePath(tt.input)
			if !tt.shouldPanic {
				if token0 != tt.wantToken0 {
					t.Errorf("token0 = %v, want %v", token0, tt.wantToken0)
				}
				if token1 != tt.wantToken1 {
					t.Errorf("token1 = %v, want %v", token1, tt.wantToken1)
				}
				if int(fee) != tt.wantFee {
					t.Errorf("fee = %v, want %v", fee, tt.wantFee)
				}
			}
		})
	}
}

func TestGetDataForMultiPath(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		poolIdx    int
		wantToken0 string
		wantToken1 string
		wantFee    uint32
	}{
		{
			name:       "first pool",
			input:      "tokenA:tokenB:500*POOL*tokenB:tokenC:3000*POOL*tokenC:tokenD:10000",
			poolIdx:    0,
			wantToken0: "tokenA",
			wantToken1: "tokenB",
			wantFee:    500,
		},
		{
			name:       "second pool",
			input:      "tokenA:tokenB:500*POOL*tokenB:tokenC:3000*POOL*tokenC:tokenD:10000",
			poolIdx:    1,
			wantToken0: "tokenB",
			wantToken1: "tokenC",
			wantFee:    3000,
		},
		{
			name:       "third pool",
			input:      "tokenA:tokenB:500*POOL*tokenB:tokenC:3000*POOL*tokenC:tokenD:10000",
			poolIdx:    2,
			wantToken0: "tokenC",
			wantToken1: "tokenD",
			wantFee:    10000,
		},
		{
			name:       "invalid pool index",
			input:      "tokenA:tokenB:500*POOL*tokenB:tokenC:3000",
			poolIdx:    3,
			wantToken0: "",
			wantToken1: "",
			wantFee:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token0, token1, fee := getDataForMultiPath(tt.input, tt.poolIdx)

			if token0 != tt.wantToken0 {
				t.Errorf("token0 = %v, want %v", token0, tt.wantToken0)
			}
			if token1 != tt.wantToken1 {
				t.Errorf("token1 = %v, want %v", token1, tt.wantToken1)
			}
			if fee != tt.wantFee {
				t.Errorf("fee = %v, want %v", fee, tt.wantFee)
			}
		})
	}
}

func TestSplitSingleChar(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		sep      byte
		expected []string
	}{
		{
			name:     "plain split",
			input:    "a,b,c",
			sep:      ',',
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty string",
			input:    "",
			sep:      ',',
			expected: []string{""},
		},
		{
			name:     "no separator",
			input:    "abc",
			sep:      ',',
			expected: []string{"abc"},
		},
		{
			name:     "consecutive separators",
			input:    "a,,b,,c",
			sep:      ',',
			expected: []string{"a", "", "b", "", "c"},
		},
		{
			name:     "separator at the beginning and end",
			input:    ",a,b,c,",
			sep:      ',',
			expected: []string{"", "a", "b", "c", ""},
		},
		{
			name:     "space separator",
			input:    "a b c",
			sep:      ' ',
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single character string",
			input:    "a",
			sep:      ',',
			expected: []string{"a"},
		},
		{
			name:     "only separators",
			input:    ",,,,",
			sep:      ',',
			expected: []string{"", "", "", "", ""},
		},
		{
			name:     "unicode characters",
			input:    "한글,English,日本語",
			sep:      ',',
			expected: []string{"한글", "English", "日本語"},
		},
		{
			name:     "special characters",
			input:    "!@#$,%^&*,()_+",
			sep:      ',',
			expected: []string{"!@#$", "%^&*", "()_+"},
		},
		{
			name:     "routes path",
			input:    "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:500*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:500,gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:500*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:500",
			sep:      ',',
			expected: []string{"gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:500*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:500", "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:500*POOL*gno.land/r/onbloc/baz:gno.land/r/onbloc/qux:500"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitSingleChar(tc.input, tc.sep)

			uassert.Equal(t, len(result), len(tc.expected))

			for i := 0; i < len(tc.expected); i++ {
				uassert.Equal(t, result[i], tc.expected[i])
			}
		})
	}
}
