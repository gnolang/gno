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

type stringer string

func (s stringer) String() string { return string(s) }

func main() {
	expect.Value(t, "foo").ToEqual("bar")
	expect.Value(t, []byte("foo")).ToEqual([]byte("bar"))
	expect.Value(t, stringer("foo")).ToEqual(stringer("bar"))
	expect.Value(t, true).ToEqual(false)
	expect.Value(t, float32(1)).ToEqual(float32(2))
	expect.Value(t, float64(1.1)).ToEqual(float64(1.2))
	expect.Value(t, uint(1)).ToEqual(uint(2))
	expect.Value(t, uint8(1)).ToEqual(uint8(2))
	expect.Value(t, uint16(1)).ToEqual(uint16(2))
	expect.Value(t, uint32(1)).ToEqual(uint32(2))
	expect.Value(t, uint64(1)).ToEqual(uint64(2))
	expect.Value(t, int(1)).ToEqual(int(2))
	expect.Value(t, int8(1)).ToEqual(int8(2))
	expect.Value(t, int16(1)).ToEqual(int16(2))
	expect.Value(t, int32(1)).ToEqual(int32(2))
	expect.Value(t, int64(1)).ToEqual(int64(2))
	expect.Value(t, errors.New("foo")).ToEqual(errors.New("bar"))
	expect.Value(t, errors.New("foo")).ToContainErrorString("bar")

	expect.Value(t, 0).ToEqual(errors.New("foo"))
	expect.Value(t, 0).ToEqual([]string{})

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
// Expected errors to match
// Got: foo
// Want: bar
// Expected error message to contain: bar
// Got: foo
// Error is not equal to value
// Got: foo
// Unsupported type: unknown
