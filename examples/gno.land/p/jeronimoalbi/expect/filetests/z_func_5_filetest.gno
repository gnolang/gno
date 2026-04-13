package main

import (
	"strings"

	"gno.land/p/jeronimoalbi/expect"
)

var (
	output strings.Builder
	t      = expect.MockTestingT(&output)
)

func main() {
	expect.Func(t, func() any { return "foo" }).Not().ToReturn("foo")
	expect.Func(t, func() any { return []byte("foo") }).Not().ToReturn([]byte("foo"))
	expect.Func(t, func() any { return true }).Not().ToReturn(true)
	expect.Func(t, func() any { return float32(1) }).Not().ToReturn(float32(1))
	expect.Func(t, func() any { return float64(1.1) }).Not().ToReturn(float64(1.1))
	expect.Func(t, func() any { return uint(1) }).Not().ToReturn(uint(1))
	expect.Func(t, func() any { return uint8(1) }).Not().ToReturn(uint8(1))
	expect.Func(t, func() any { return uint16(1) }).Not().ToReturn(uint16(1))
	expect.Func(t, func() any { return uint32(1) }).Not().ToReturn(uint32(1))
	expect.Func(t, func() any { return uint64(1) }).Not().ToReturn(uint64(1))
	expect.Func(t, func() any { return int(1) }).Not().ToReturn(int(1))
	expect.Func(t, func() any { return int8(1) }).Not().ToReturn(int8(1))
	expect.Func(t, func() any { return int16(1) }).Not().ToReturn(int16(1))
	expect.Func(t, func() any { return int32(1) }).Not().ToReturn(int32(1))
	expect.Func(t, func() any { return int64(1) }).Not().ToReturn(int64(1))

	expect.Func(t, func() (any, error) { return "foo", nil }).Not().ToReturn("foo")
	expect.Func(t, func() (any, error) { return []byte("foo"), nil }).Not().ToReturn([]byte("foo"))
	expect.Func(t, func() (any, error) { return true, nil }).Not().ToReturn(true)
	expect.Func(t, func() (any, error) { return float32(1), nil }).Not().ToReturn(float32(1))
	expect.Func(t, func() (any, error) { return float64(1.1), nil }).Not().ToReturn(float64(1.1))
	expect.Func(t, func() (any, error) { return uint(1), nil }).Not().ToReturn(uint(1))
	expect.Func(t, func() (any, error) { return uint8(1), nil }).Not().ToReturn(uint8(1))
	expect.Func(t, func() (any, error) { return uint16(1), nil }).Not().ToReturn(uint16(1))
	expect.Func(t, func() (any, error) { return uint32(1), nil }).Not().ToReturn(uint32(1))
	expect.Func(t, func() (any, error) { return uint64(1), nil }).Not().ToReturn(uint64(1))
	expect.Func(t, func() (any, error) { return int(1), nil }).Not().ToReturn(int(1))
	expect.Func(t, func() (any, error) { return int8(1), nil }).Not().ToReturn(int8(1))
	expect.Func(t, func() (any, error) { return int16(1), nil }).Not().ToReturn(int16(1))
	expect.Func(t, func() (any, error) { return int32(1), nil }).Not().ToReturn(int32(1))
	expect.Func(t, func() (any, error) { return int64(1), nil }).Not().ToReturn(int64(1))

	println(output.String())
}

// Output:
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: true
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1.1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: true
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1.1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
