package unsafe

import (
	"errors"
	"math"
	"strconv"

	"gno.land/p/demo/ufmt"
)

func Sizeof(x any) uint {
	switch x.(type) {
	// case complex128:
	// 	return 16
	case float64, uint64, int64 /*, complex64 */:
		return 8
	case uint32, int32, float32: // + rune
		return 4
	case uint16, int16:
		return 2
	case int8, uint8, bool: // + byte
		return 1
	case struct{}:
		return 0

	// UNSURE: this types depends on architecture it has to be changed
	case string:
		return 8 + 8
	case /*uintptr,*/ int, uint:
		return 8
	case *float64, *uint64, *int64, /* *complex64 , *complex128,*/ *uint32, *int32, *float32, *int16, *uint16, *int8, *uint8, *bool, *string, /* *uintptr,*/ *int, *uint:
		return 8
	case []float64, []uint64, []int64, /* []complex64 , []complex128,*/ []uint32, []int32, []float32, []int8, []uint8, []bool, []string, /*[]uintptr,*/ []int, []uint:
		return 8 + 8 + 8
	case map[any]any:
		return 24
	case chan any:
		return 8 + 8
	}
	return 0
}

func ObjectToBits(x any) (uint64, error) {
	switch v := x.(type) {
	case uint64:
		return v, nil
	case uint32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case float32:
		return uint64(math.Float32bits(v)), nil
	case float64:
		return math.Float64bits(v), nil
	case *float64, *uint64, *int64, /* *complex64, *complex128,*/ *uint32, *int32, *float32, *int16, *uint16, *int8, *uint8, *bool, *string, /* *uintptr,*/ *int, *uint, *map[any]any, *struct{}, *chan any:
		n, err := strconv.ParseUint(ufmt.Sprintf("%p", v), 0, 64)
		if (err != nil) {
			return 0, err
		}
		return n, nil
	}
	return 0, errors.New("cant convert type")
}

func WriteBytes(x any) ([]byte, error) {
	bits, err := ObjectToBits(x)
	if err != nil {
		return []byte{}, err
	}
	byteCount := Sizeof(x)
	result := make([]byte, byteCount)
	for i := uint(0); i < byteCount; i++ {
		result[i] = byte(bits >> ((byteCount-1-i)*8))
	}
	return result, nil
}