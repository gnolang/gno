package vm

import (
	"encoding/base64"
	"fmt"
	"math"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertEmptyNumbers(t *testing.T) {
	tests := []struct {
		argT        gnolang.Type
		expectedErr string
	}{
		{gnolang.UintType, `error parsing uint "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint64Type, `error parsing uint64 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint32Type, `error parsing uint32 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint16Type, `error parsing uint16 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.Uint8Type, `error parsing uint8 "": strconv.ParseUint: parsing "": invalid syntax`},
		{gnolang.IntType, `error parsing int "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int64Type, `error parsing int64 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int32Type, `error parsing int32 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int16Type, `error parsing int16 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Int8Type, `error parsing int8 "": strconv.ParseInt: parsing "": invalid syntax`},
		{gnolang.Float64Type, `error parsing float64 "": parse mantissa: `},
		{gnolang.Float32Type, `error parsing float32 "": parse mantissa: `},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.argT)
		t.Run(testname, func(t *testing.T) {
			run := func() {
				_ = convertArgToGno("", tt.argT)
			}
			assert.PanicsWithValue(t, tt.expectedErr, run)
		})
	}
}

func TestConvertByteArrayLengthValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		declaredLen int
		inputLen    int
		shouldPanic bool
	}{
		{"exact match", 32, 32, false},
		{"oversized input", 32, 100, true},
		{"undersized input", 32, 16, true},
		{"empty input", 32, 0, true},
		{"one byte array exact", 1, 1, false},
		{"one byte array oversized", 1, 2, true},
		{"zero length array exact", 0, 0, false},
		{"zero length array oversized", 0, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			arrType := &gnolang.ArrayType{Len: tt.declaredLen, Elt: gnolang.Uint8Type}
			input := make([]byte, tt.inputLen)
			for i := range input {
				input[i] = byte(i)
			}
			b64 := base64.StdEncoding.EncodeToString(input)

			if tt.shouldPanic {
				require.PanicsWithValue(t, fmt.Sprintf("array length mismatch: declared [%d]byte, got %d bytes", tt.declaredLen, tt.inputLen), func() {
					convertArgToGno(b64, arrType)
				})
			} else {
				tv := convertArgToGno(b64, arrType)
				av, ok := tv.V.(*gnolang.ArrayValue)
				require.True(t, ok)
				assert.Equal(t, tt.declaredLen, av.GetLength())
			}
		})
	}
}

func TestConvertFloatInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		arg         string
		argT        gnolang.Type
		expectedErr string
	}{
		// NaN
		{"float64 NaN", "NaN", gnolang.Float64Type, "float64 does not accept NaN"},
		{"float32 NaN", "NaN", gnolang.Float32Type, "float32 does not accept NaN"},
		// Inf
		{"float64 Inf", "Inf", gnolang.Float64Type, "float64 does not accept Inf"},
		{"float32 Inf", "Inf", gnolang.Float32Type, "float32 does not accept Inf"},
		{"float64 -Inf", "-Inf", gnolang.Float64Type, "float64 does not accept Inf"},
		{"float32 -Inf", "-Inf", gnolang.Float32Type, "float32 does not accept Inf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.PanicsWithValue(t, tt.expectedErr, func() {
				convertArgToGno(tt.arg, tt.argT)
			})
		})
	}
}

func TestConvertFloatNegativeZeroCanonicalized(t *testing.T) {
	t.Parallel()

	// -0.0 and -0 must both produce positive zero bits to prevent malleability.
	for _, arg := range []string{"-0.0", "-0"} {
		t.Run(arg, func(t *testing.T) {
			t.Parallel()

			tv64 := convertArgToGno(arg, gnolang.Float64Type)
			bits64 := tv64.GetFloat64()
			require.False(t, math.Signbit(math.Float64frombits(bits64)),
				"float64 %q: expected positive zero bits, got 0x%016X", arg, bits64)

			tv32 := convertArgToGno(arg, gnolang.Float32Type)
			bits32 := tv32.GetFloat32()
			require.False(t, math.Signbit(float64(math.Float32frombits(bits32))),
				"float32 %q: expected positive zero bits, got 0x%08X", arg, bits32)
		})
	}
}
