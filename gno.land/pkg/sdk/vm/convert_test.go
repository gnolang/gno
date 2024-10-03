package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno/pkg/vm"
	"github.com/stretchr/testify/assert"
)

func TestConvertEmptyNumbers(t *testing.T) {
	tests := []struct {
		argT        vm.Type
		expectedErr string
	}{
		{vm.UintType, `error parsing uint "": strconv.ParseUint: parsing "": invalid syntax`},
		{vm.Uint64Type, `error parsing uint64 "": strconv.ParseUint: parsing "": invalid syntax`},
		{vm.Uint32Type, `error parsing uint32 "": strconv.ParseUint: parsing "": invalid syntax`},
		{vm.Uint16Type, `error parsing uint16 "": strconv.ParseUint: parsing "": invalid syntax`},
		{vm.Uint8Type, `error parsing uint8 "": strconv.ParseUint: parsing "": invalid syntax`},
		{vm.IntType, `error parsing int "": strconv.ParseInt: parsing "": invalid syntax`},
		{vm.Int64Type, `error parsing int64 "": strconv.ParseInt: parsing "": invalid syntax`},
		{vm.Int32Type, `error parsing int32 "": strconv.ParseInt: parsing "": invalid syntax`},
		{vm.Int16Type, `error parsing int16 "": strconv.ParseInt: parsing "": invalid syntax`},
		{vm.Int8Type, `error parsing int8 "": strconv.ParseInt: parsing "": invalid syntax`},
		{vm.Float64Type, `error parsing float64 "": parse mantissa: `},
		{vm.Float32Type, `error parsing float32 "": parse mantissa: `},
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
