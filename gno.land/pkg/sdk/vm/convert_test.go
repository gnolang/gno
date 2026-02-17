package vm

import (
	"encoding/base64"
	"fmt"
	"strings"
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
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					convertArgToGno(b64, arrType)
				}()
				require.NotNil(t, recovered, "expected panic for [%d]byte with %d bytes input", tt.declaredLen, tt.inputLen)
				require.True(t, strings.Contains(fmt.Sprint(recovered), "array length mismatch"),
					"expected 'array length mismatch' in panic, got: %v", recovered)
			} else {
				tv := convertArgToGno(b64, arrType)
				av, ok := tv.V.(*gnolang.ArrayValue)
				require.True(t, ok)
				assert.Equal(t, tt.declaredLen, av.GetLength())
			}
		})
	}
}
