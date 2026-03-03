package main

import (
	"errors"
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func main() {
	expect.Func(t, func() any { return "foo" }).ToReturn("bar")
	expect.Func(t, func() any { return []byte("foo") }).ToReturn([]byte("bar"))
	expect.Func(t, func() any { return true }).ToReturn(false)
	expect.Func(t, func() any { return float32(1) }).ToReturn(float32(2))
	expect.Func(t, func() any { return float64(1.1) }).ToReturn(float64(1.2))
	expect.Func(t, func() any { return uint(1) }).ToReturn(uint(2))
	expect.Func(t, func() any { return uint8(1) }).ToReturn(uint8(2))
	expect.Func(t, func() any { return uint16(1) }).ToReturn(uint16(2))
	expect.Func(t, func() any { return uint32(1) }).ToReturn(uint32(2))
	expect.Func(t, func() any { return uint64(1) }).ToReturn(uint64(2))
	expect.Func(t, func() any { return int(1) }).ToReturn(int(2))
	expect.Func(t, func() any { return int8(1) }).ToReturn(int8(2))
	expect.Func(t, func() any { return int16(1) }).ToReturn(int16(2))
	expect.Func(t, func() any { return int32(1) }).ToReturn(int32(2))
	expect.Func(t, func() any { return int64(1) }).ToReturn(int64(2))

	expect.Func(t, func() (any, error) { return "foo", nil }).ToReturn("bar")
	expect.Func(t, func() (any, error) { return []byte("foo"), nil }).ToReturn([]byte("bar"))
	expect.Func(t, func() (any, error) { return true, nil }).ToReturn(false)
	expect.Func(t, func() (any, error) { return float32(1), nil }).ToReturn(float32(2))
	expect.Func(t, func() (any, error) { return float64(1.1), nil }).ToReturn(float64(1.2))
	expect.Func(t, func() (any, error) { return uint(1), nil }).ToReturn(uint(2))
	expect.Func(t, func() (any, error) { return uint8(1), nil }).ToReturn(uint8(2))
	expect.Func(t, func() (any, error) { return uint16(1), nil }).ToReturn(uint16(2))
	expect.Func(t, func() (any, error) { return uint32(1), nil }).ToReturn(uint32(2))
	expect.Func(t, func() (any, error) { return uint64(1), nil }).ToReturn(uint64(2))
	expect.Func(t, func() (any, error) { return int(1), nil }).ToReturn(int(2))
	expect.Func(t, func() (any, error) { return int8(1), nil }).ToReturn(int8(2))
	expect.Func(t, func() (any, error) { return int16(1), nil }).ToReturn(int16(2))
	expect.Func(t, func() (any, error) { return int32(1), nil }).ToReturn(int32(2))
	expect.Func(t, func() (any, error) { return int64(1), nil }).ToReturn(int64(2))

	expect.Func(t, func() (any, error) { return 0, errors.New("Boom!") }).ToReturn(1)
	expect.Func(t, func() {}).ToReturn(1)

	println(output.String())
}

// Output:
// Expected values to match
// Got: foo
// Want: bar
// Expected values to match
// Got: foo
// Want: bar
// Expected values to match
// Got: true
// Want: false
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1.1
// Want: 1.2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: foo
// Want: bar
// Expected values to match
// Got: foo
// Want: bar
// Expected values to match
// Got: true
// Want: false
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1.1
// Want: 1.2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Expected values to match
// Got: 1
// Want: 2
// Function returned unexpected error
// Got: Boom!
// Unsupported func type
// Got: unknown
