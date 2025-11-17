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
	expect.Value(t, "foo").Not().ToEqual("foo")
	expect.Value(t, []byte("foo")).Not().ToEqual([]byte("foo"))
	expect.Value(t, stringer("foo")).Not().ToEqual(stringer("foo"))
	expect.Value(t, true).Not().ToEqual(true)
	expect.Value(t, float32(1)).Not().ToEqual(float32(1))
	expect.Value(t, float64(1)).Not().ToEqual(float64(1))
	expect.Value(t, uint(1)).Not().ToEqual(uint(1))
	expect.Value(t, uint8(1)).Not().ToEqual(uint8(1))
	expect.Value(t, uint16(1)).Not().ToEqual(uint16(1))
	expect.Value(t, uint32(1)).Not().ToEqual(uint32(1))
	expect.Value(t, uint64(1)).Not().ToEqual(uint64(1))
	expect.Value(t, int(1)).Not().ToEqual(int(1))
	expect.Value(t, int8(1)).Not().ToEqual(int8(1))
	expect.Value(t, int16(1)).Not().ToEqual(int16(1))
	expect.Value(t, int32(1)).Not().ToEqual(int32(1))
	expect.Value(t, int64(1)).Not().ToEqual(int64(1))
	expect.Value(t, errors.New("foo")).Not().ToEqual(errors.New("foo"))
	expect.Value(t, errors.New("foo bar")).Not().ToContainErrorString("bar")

	expect.Value(t, 0).Not().ToEqual(errors.New("foo"))
	expect.Value(t, 0).Not().ToEqual([]string{})

	println(output.String())
}

// Output:
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: foo
// Expected values to be different
// Got: true
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
// Expected value to be different
// Got: 1
// Expected value to be different
// Got: 1
// Expected errors to be different
// Got: foo
// Expected error message not to contain: bar
// Got: foo bar
// Error is not equal to value
// Got: foo
// Unsupported type: unknown
